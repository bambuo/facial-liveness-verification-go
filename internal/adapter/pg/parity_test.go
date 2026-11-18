package pg

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type columnInfo struct {
	Type     string
	Nullable string
	Default  string
}

type snapshot struct {
	Tables  []string
	Columns map[string]map[string]columnInfo
	PKs     map[string][]string
	FKs     []string
	Checks  []string
	Indexes []string
}

// TestSchemaParity 对比 AutoMigrate 与 001_baseline.sql 的语义 schema。
func TestSchemaParity(t *testing.T) {
	gormURL := os.Getenv("TEST_GORM_URL")
	baselineURL := os.Getenv("TEST_BASELINE_URL")
	if gormURL == "" || baselineURL == "" {
		t.Skip("TEST_GORM_URL / TEST_BASELINE_URL 未设置")
	}
	gormDB := openSQL(t, gormURL)
	baselineDB := openSQL(t, baselineURL)
	want := takeSnapshot(t, baselineDB)
	got := takeSnapshot(t, gormDB)

	if !reflect.DeepEqual(want.Tables, got.Tables) {
		t.Errorf("tables differ:\nwant=%v\ngot=%v", want.Tables, got.Tables)
	}
	if !reflect.DeepEqual(want.Columns, got.Columns) {
		for table, cols := range want.Columns {
			gcols := got.Columns[table]
			for name, c := range cols {
				if g, ok := gcols[name]; !ok || g != c {
					t.Errorf("column %s.%s: want=%+v got=%+v", table, name, c, g)
				}
			}
			for name := range gcols {
				if _, ok := cols[name]; !ok {
					t.Errorf("extra column %s.%s", table, name)
				}
			}
		}
	}
	if !reflect.DeepEqual(want.PKs, got.PKs) {
		t.Errorf("primary keys differ:\nwant=%v\ngot=%v", want.PKs, got.PKs)
	}
	if !reflect.DeepEqual(want.FKs, got.FKs) {
		t.Errorf("foreign keys differ:\nwant=%v\ngot=%v", want.FKs, got.FKs)
	}
	if !reflect.DeepEqual(want.Checks, got.Checks) {
		t.Errorf("checks differ:\nwant=%v\ngot=%v", want.Checks, got.Checks)
	}
	if !reflect.DeepEqual(want.Indexes, got.Indexes) {
		t.Errorf("indexes differ:\nwant=%v\ngot=%v", want.Indexes, got.Indexes)
	}
}

func openSQL(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func takeSnapshot(t *testing.T, db *sql.DB) snapshot {
	t.Helper()
	s := snapshot{
		Columns: map[string]map[string]columnInfo{},
		PKs:     map[string][]string{},
	}
	rows, err := db.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		s.Tables = append(s.Tables, name)
	}
	rows.Close()

	rows, err = db.Query(`
SELECT table_name, column_name, data_type, udt_name, is_nullable, COALESCE(column_default,''),
       COALESCE(character_maximum_length,0), COALESCE(numeric_precision,0), COALESCE(numeric_scale,0)
FROM information_schema.columns WHERE table_schema='public' ORDER BY table_name, column_name`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var table, column, dataType, udt, nullable, def string
		var charLen, precision, scale int
		if err := rows.Scan(&table, &column, &dataType, &udt, &nullable, &def, &charLen, &precision, &scale); err != nil {
			t.Fatal(err)
		}
		typ := udt
		switch udt {
		case "varchar", "bpchar":
			typ = fmt.Sprintf("%s(%d)", udt, charLen)
		case "numeric":
			typ = fmt.Sprintf("numeric(%d,%d)", precision, scale)
		}
		if s.Columns[table] == nil {
			s.Columns[table] = map[string]columnInfo{}
		}
		s.Columns[table][column] = columnInfo{Type: typ, Nullable: nullable, Default: normalizeDefault(def)}
	}
	rows.Close()

	rows, err = db.Query(`
SELECT c.relname, a.attname, array_position(con.conkey, a.attnum)
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(con.conkey)
WHERE con.contype='p' AND c.relnamespace='public'::regnamespace
ORDER BY c.relname, 3`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var table, column string
		var pos int
		if err := rows.Scan(&table, &column, &pos); err != nil {
			t.Fatal(err)
		}
		s.PKs[table] = append(s.PKs[table], column)
	}
	rows.Close()

	rows, err = db.Query(`
SELECT conrelid::regclass::text, pg_get_constraintdef(oid)
FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var table, def string
		if err := rows.Scan(&table, &def); err != nil {
			t.Fatal(err)
		}
		table = strings.TrimPrefix(table, "public.")
		s.FKs = append(s.FKs, table+"|"+normalizeExpr(def))
	}
	rows.Close()

	rows, err = db.Query(`
SELECT conrelid::regclass::text, pg_get_constraintdef(oid)
FROM pg_constraint WHERE contype='c' AND connamespace='public'::regnamespace`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var table, def string
		if err := rows.Scan(&table, &def); err != nil {
			t.Fatal(err)
		}
		table = strings.TrimPrefix(table, "public.")
		s.Checks = append(s.Checks, table+"|"+normalizeExpr(def))
	}
	rows.Close()

	rows, err = db.Query(`SELECT tablename, indexdef FROM pg_indexes WHERE schemaname='public'`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var table, def string
		if err := rows.Scan(&table, &def); err != nil {
			t.Fatal(err)
		}
		s.Indexes = append(s.Indexes, table+"|"+normalizeIndex(def))
	}
	rows.Close()

	sort.Strings(s.FKs)
	sort.Strings(s.Checks)
	sort.Strings(s.Indexes)
	return s
}

var (
	castRe  = regexp.MustCompile(`::[a-zA-Z_ ]+(\(\d+(,\d+)?\))?`)
	spaceRe = regexp.MustCompile(`\s+`)
)

func normalizeDefault(def string) string {
	if def == "" {
		return ""
	}
	d := strings.ToLower(def)
	d = castRe.ReplaceAllString(d, "")
	d = strings.ReplaceAll(d, "current_timestamp", "now()")
	d = spaceRe.ReplaceAllString(d, "")
	return d
}

func normalizeExpr(def string) string {
	d := strings.ToLower(def)
	d = castRe.ReplaceAllString(d, "")
	d = spaceRe.ReplaceAllString(d, "")
	for strings.HasPrefix(d, "((") && strings.HasSuffix(d, "))") {
		d = d[1 : len(d)-1]
	}
	return d
}

func normalizeIndex(def string) string {
	d := strings.ToLower(def)
	d = strings.ReplaceAll(d, "public.", "")
	d = regexp.MustCompile(`create\s+(unique\s+)?index\s+\S+\s+on\s+`).ReplaceAllString(d, "create ${1}index on ")
	d = regexp.MustCompile(`using\s+btree\s*`).ReplaceAllString(d, "")
	d = castRe.ReplaceAllString(d, "")
	d = spaceRe.ReplaceAllString(d, "")
	for strings.HasPrefix(d, "((") && strings.HasSuffix(d, "))") {
		d = d[1 : len(d)-1]
	}
	return strings.TrimSpace(d)
}
