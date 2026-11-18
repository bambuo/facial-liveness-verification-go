// 契约校验命令：检查冻结契约的数量、鉴权签字与实现覆盖。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Module string `json:"module"`
	Auth   string `json:"auth"`
}

type endpointsDoc struct {
	Statistics struct {
		RouteRegistrations int            `json:"routeRegistrations"`
		ExpandedMethodPath int            `json:"expandedMethodPath"`
		ByModule           map[string]int `json:"byModule"`
	} `json:"statistics"`
	Endpoints []endpoint `json:"endpoints"`
}

type schemaDoc struct {
	Statistics struct {
		Tables  int `json:"tables"`
		Columns int `json:"columns"`
	} `json:"statistics"`
}

type fieldSpec struct {
	Name       string   `json:"name"`
	Validators []string `json:"validators"`
}

type endpointFields struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Module string      `json:"module"`
	Body   []fieldSpec `json:"body"`
	Query  []fieldSpec `json:"query"`
	Param  []fieldSpec `json:"param"`
	Files  []fieldSpec `json:"files"`
}

type fieldsDoc struct {
	Statistics struct {
		ZodSchemas int `json:"zodSchemas"`
	} `json:"statistics"`
	Endpoints []endpointFields `json:"endpoints"`
}

type responseRecord struct {
	Status string              `json:"status"`
	Keys   []string            `json:"keys"`
	Nested map[string][]string `json:"nested"`
	Ref    string              `json:"ref"`
	Kind   string              `json:"kind"`
}

type responseDoc struct {
	Endpoints []struct {
		Method    string           `json:"method"`
		Path      string           `json:"path"`
		Responses []responseRecord `json:"responses"`
	} `json:"endpoints"`
}

type fixture struct {
	Method      string           `json:"method"`
	Path        string           `json:"path"`
	Module      string           `json:"module"`
	Auth        string           `json:"auth"`
	Implemented bool             `json:"implemented"`
	Request     map[string]any   `json:"request"`
	Responses   []responseRecord `json:"responses"`
}

type engineDoc struct {
	Endpoints []endpoint `json:"endpoints"`
}

// main 执行契约校验并生成覆盖报告。
func main() {
	dir := flag.String("contracts", "contracts", "契约目录")
	handlersPath := flag.String("handlers", "internal/adapter/web/handlers.go", "路由源码")
	out := flag.String("out", "contracts/coverage.md", "覆盖报告输出")
	flag.Parse()

	var failures []string
	fail := func(format string, args ...any) { failures = append(failures, fmt.Sprintf(format, args...)) }

	eps := loadJSON[endpointsDoc](filepath.Join(*dir, "endpoints.json"))
	sch := loadJSON[schemaDoc](filepath.Join(*dir, "schema.json"))
	fld := loadJSON[fieldsDoc](filepath.Join(*dir, "fields.json"))
	auth := loadJSON[endpointsDoc](filepath.Join(*dir, "auth.json"))
	resp := loadJSON[responseDoc](filepath.Join(*dir, "responses.json"))
	eng := loadJSON[engineDoc](filepath.Join(*dir, "engine-endpoints.json"))
	if len(eng.Endpoints) != 6 {
		fail("引擎端点 %d != 6", len(eng.Endpoints))
	}

	if eps.Statistics.RouteRegistrations != 154 {
		fail("接口注册数 %d != 154", eps.Statistics.RouteRegistrations)
	}
	if sch.Statistics.Tables != 31 || sch.Statistics.Columns != 329 {
		fail("表/列 %d/%d != 31/329", sch.Statistics.Tables, sch.Statistics.Columns)
	}
	if fld.Statistics.ZodSchemas != 16 {
		fail("Zod schema %d != 16", fld.Statistics.ZodSchemas)
	}

	authMap := map[string]endpoint{}
	for _, e := range auth.Endpoints {
		authMap[key(e.Method, e.Path)] = e
	}
	fieldsMap := map[string]endpointFields{}
	for _, e := range fld.Endpoints {
		fieldsMap[key(e.Method, e.Path)] = e
	}
	respMap := map[string][]responseRecord{}
	for _, e := range resp.Endpoints {
		respMap[key(e.Method, e.Path)] = e.Responses
	}
	signed := loadSigned(filepath.Join(*dir, "auth-decisions.md"))
	publicCandidates := 0
	for _, e := range eps.Endpoints {
		k := key(e.Method, e.Path)
		a, ok := authMap[k]
		if !ok {
			fail("缺少鉴权证据: %s", k)
			continue
		}
		if a.Auth == "public-candidate" {
			publicCandidates++
			if !signed[k] {
				fail("公开候选未签字: %s", k)
			}
		}
		if len(respMap[k]) == 0 {
			fail("缺少响应证据: %s", k)
		}
	}
	if publicCandidates != 17 {
		fail("公开候选数 %d != 17", publicCandidates)
	}
	for k := range signed {
		if _, ok := authMap[k]; !ok {
			fail("签字路径不在鉴权清单: %s", k)
		}
	}

	implemented := parseHandlers(*handlersPath)
	fixtures := make([]fixture, 0, len(eps.Endpoints))
	nestedCount := 0
	for _, e := range eps.Endpoints {
		k := key(e.Method, e.Path)
		responses := respMap[k]
		hasEvidence := false
		for _, r := range responses {
			if strings.HasPrefix(r.Status, "2") && len(r.Keys) == 0 && r.Ref == "" {
				fail("2xx 响应缺少字段证据: %s", k)
			}
			if len(r.Keys) > 0 || r.Ref != "" || strings.HasPrefix(r.Status, "2") {
				hasEvidence = true
			}
			if len(r.Nested) > 0 {
				nestedCount++
			}
		}
		if !hasEvidence {
			fail("缺少响应字段证据: %s", k)
		}
		req := fieldsMap[k]
		authValue := ""
		if a, ok := authMap[k]; ok {
			authValue = a.Auth
		}
		fixtures = append(fixtures, fixture{
			Method:      e.Method,
			Path:        e.Path,
			Module:      e.Module,
			Auth:        authValue,
			Implemented: implemented[k] || implemented[key(e.Method, normalize(e.Path))],
			Request: map[string]any{
				"body": req.Body, "query": req.Query, "param": req.Param, "files": req.Files,
			},
			Responses: responses,
		})
	}
	fixtureDoc := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"source":      "contracts/endpoints.json + responses.json + fields.json",
		"endpoints":   fixtures,
	}
	fixtureRaw, err := json.MarshalIndent(fixtureDoc, "", "  ")
	if err != nil {
		fail("生成 fixtures: %v", err)
	} else if err := os.WriteFile(filepath.Join(*dir, "fixtures.json"), append(fixtureRaw, '\n'), 0o644); err != nil {
		fail("写入 fixtures: %v", err)
	}

	matched := make([]string, 0)
	for _, e := range eps.Endpoints {
		if implemented[key(e.Method, normalize(e.Path))] {
			matched = append(matched, key(e.Method, e.Path))
		}
	}
	sort.Strings(matched)

	engineMatched := make([]string, 0)
	for _, e := range eng.Endpoints {
		if implemented[key(e.Method, normalize(e.Path))] {
			engineMatched = append(engineMatched, key(e.Method, e.Path))
		}
	}
	sort.Strings(engineMatched)

	var b strings.Builder
	b.WriteString("# 契约覆盖报告\n\n")
	b.WriteString("由 `go run ./cmd/contractcheck` 生成。\n\n")
	b.WriteString("## 冻结契约\n\n")
	b.WriteString("| 项 | 数值 |\n| --- | --- |\n")
	fmt.Fprintf(&b, "| 接口注册 | %d |\n", eps.Statistics.RouteRegistrations)
	fmt.Fprintf(&b, "| 可达路径 | %d |\n", eps.Statistics.ExpandedMethodPath)
	fmt.Fprintf(&b, "| 数据表 / 列 | %d / %d |\n", sch.Statistics.Tables, sch.Statistics.Columns)
	fmt.Fprintf(&b, "| Zod schema | %d |\n", fld.Statistics.ZodSchemas)
	fmt.Fprintf(&b, "| 契约 fixtures | %d（含嵌套响应 %d） |\n", len(fixtures), nestedCount)
	fmt.Fprintf(&b, "| 公开候选签字 | %d |\n", publicCandidates)
	b.WriteString("\n## 实现覆盖\n\n")
	fmt.Fprintf(&b, "后端接口：**%d / %d** 可达路径（154 注册）\n\n", len(matched), len(eps.Endpoints))
	fmt.Fprintf(&b, "引擎端点：**%d / %d**\n\n", len(engineMatched), len(eng.Endpoints))
	if len(matched) > 0 {
		b.WriteString("### 已实现后端接口\n\n| 方法 | 路径 |\n| --- | --- |\n")
		for _, m := range matched {
			parts := strings.SplitN(m, " ", 2)
			fmt.Fprintf(&b, "| %s | `%s` |\n", parts[0], parts[1])
		}
		b.WriteString("\n")
	}
	if len(engineMatched) > 0 {
		b.WriteString("### 已实现引擎端点\n\n| 方法 | 路径 |\n| --- | --- |\n")
		for _, m := range engineMatched {
			parts := strings.SplitN(m, " ", 2)
			fmt.Fprintf(&b, "| %s | `%s` |\n", parts[0], parts[1])
		}
		b.WriteString("\n")
	}
	b.WriteString("## 待实现模块\n\n| 模块 | 接口数 |\n| --- | --- |\n")
	modules := make([]string, 0, len(eps.Statistics.ByModule))
	for m := range eps.Statistics.ByModule {
		modules = append(modules, m)
	}
	sort.Strings(modules)
	for _, m := range modules {
		fmt.Fprintf(&b, "| %s | %d |\n", m, eps.Statistics.ByModule[m])
	}
	b.WriteString("\n## 校验结果\n\n")
	if len(failures) == 0 {
		b.WriteString("全部通过。\n")
	} else {
		for _, f := range failures {
			fmt.Fprintf(&b, "- %s\n", f)
		}
	}

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintln(os.Stderr, "FAIL:", f)
		}
		os.Exit(1)
	}
	fmt.Printf("contract ok: %d backend endpoints, %d backend implemented, %d/%d engine implemented, report=%s\n",
		len(eps.Endpoints), len(matched), len(engineMatched), len(eng.Endpoints), *out)
}

func loadJSON[T any](path string) T {
	var out T
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", path, err)
		os.Exit(1)
	}
	return out
}

func key(method, path string) string { return method + " " + path }

var braceParam = regexp.MustCompile(`\{([^}]+)\}`)

func normalize(path string) string { return braceParam.ReplaceAllString(path, ":$1") }

var (
	handleRe = regexp.MustCompile(`mux\.HandleFunc\("([A-Z]+) ([^"]+)"`)
	rowRe    = regexp.MustCompile("\\|\\s*\\d+\\s*\\|[^|]*\\|\\s*([A-Z]+)\\s*\\|\\s*`([^`]+)`\\s*\\|")
)

// parseHandlers 从路由源码提取已注册的 method+path。
func parseHandlers(path string) map[string]bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	out := map[string]bool{}
	for _, m := range handleRe.FindAllStringSubmatch(string(raw), -1) {
		out[key(m[1], normalize(m[2]))] = true
	}
	return out
}

// loadSigned 读取签字表，返回 method+path 集合。
func loadSigned(path string) map[string]bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]bool{}
	}
	out := map[string]bool{}
	for _, m := range rowRe.FindAllStringSubmatch(string(raw), -1) {
		out[key(m[1], m[2])] = true
	}
	return out
}
