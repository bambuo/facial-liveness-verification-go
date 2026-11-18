package fields

import (
	"strings"
	"testing"
)

func padMRZ(s string) string {
	if len(s) >= 44 {
		return s[:44]
	}
	return s + strings.Repeat("<", 44-len(s))
}

// TestExtractUS 覆盖美国驾照标签与 MM/DD/YYYY 日期。
func TestExtractUS(t *testing.T) {
	text := "NAME: JOHN DOE\nDOB: 01/02/1990\nDL: D1234567\nEXP: 01/02/2030\n"
	got := Extract(text, "us")
	if got.FullName != "JOHN DOE" || got.DateOfBirth != "1990-01-02" ||
		got.IDNumber != "D1234567" || got.ExpiryDate != "2030-01-02" || got.IssuingCountry != "US" {
		t.Fatalf("unexpected: %+v", got)
	}
}

// TestExtractUK 覆盖英国驾照 16 位证号与 DD/MM/YYYY 日期。
func TestExtractUK(t *testing.T) {
	text := "NAME: JANE SMITH\nDOB: 02/03/1985\nDL: SMITH802030JA9AB\nEXP: 02/03/2035\n"
	got := Extract(text, "gb")
	if got.FullName != "JANE SMITH" || got.DateOfBirth != "1985-03-02" ||
		got.IDNumber != "SMITH802030JA9AB" || got.ExpiryDate != "2035-03-02" {
		t.Fatalf("unexpected: %+v", got)
	}
}

// TestExtractCN 覆盖中国身份证：18 位号码与出生日期。
func TestExtractCN(t *testing.T) {
	text := "姓名：张三\n出生日期：1990-01-01\n公民身份号码：110105199001011234\n有效期限：2030-01-01\n"
	got := Extract(text, "cn")
	if got.FullName != "张三" || got.DateOfBirth != "1990-01-01" ||
		got.IDNumber != "110105199001011234" || got.ExpiryDate != "2030-01-01" {
		t.Fatalf("unexpected: %+v", got)
	}
}

// TestExtractMRZ 覆盖护照 TD3 MRZ 优先解析。
func TestExtractMRZ(t *testing.T) {
	line1 := padMRZ("P<USADOE<<JOHN")
	line2 := padMRZ("E123456789USA9001011M300101")
	got := Extract(line1+"\n"+line2+"\n", "")
	if got.FullName != "JOHN DOE" || got.DateOfBirth != "1990-01-01" ||
		got.IDNumber != "E12345678" || got.ExpiryDate != "2030-01-01" {
		t.Fatalf("unexpected: %+v", got)
	}
}

// TestExtractEmpty 验证无有效字段时返回空结果。
func TestExtractEmpty(t *testing.T) {
	got := Extract("hello world", "")
	if got.FullName != "" || got.DateOfBirth != "" || got.IDNumber != "" || got.ExpiryDate != "" {
		t.Fatalf("expected empty, got %+v", got)
	}
}

// TestExtractBRLabels 覆盖巴西护照可视标签字段。
func TestExtractBRLabels(t *testing.T) {
	text := "SOBRENOME / SURNAME\nSILVA COSTA\nNOME / GIVEN NAMES\nJOANA\n" +
		"DATA DE NASCIMENTO / DATE OF BIRTH\n10 JAN/JAN 1981\n" +
		"VALIDO ATÉ / DATE OF EXPIRY\n14 AGO/AUG 2025\n" +
		"PASSAPORTE Nº / PASSPORT NO\nAA000000\n"
	got := Extract(text, "br")
	if got.FullName != "JOANA SILVA COSTA" || got.DateOfBirth != "1981-01-10" ||
		got.IDNumber != "AA000000" || got.ExpiryDate != "2025-08-14" {
		t.Fatalf("unexpected: %+v", got)
	}
}
