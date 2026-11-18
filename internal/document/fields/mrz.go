package fields

import "strings"

// MRZ 是 TD3 护照 MRZ 的结构化结果。
type MRZ struct {
	RawLines       []string
	Fields         Result
	ChecksumsValid bool
	Format         string
}

// ExtractMRZ 从 OCR 文本中查找并解析 TD3 两行 MRZ，返回原始行与校验位结果。
func ExtractMRZ(text string) (MRZ, bool) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := 0; i+1 < len(lines); i++ {
		a := mrzCandidate(lines[i], true)
		b := mrzCandidate(lines[i+1], false)
		if a == "" || b == "" {
			continue
		}
		fields := parseMRZPair(a, b)
		return MRZ{
			RawLines:       []string{strings.TrimSpace(lines[i]), strings.TrimSpace(lines[i+1])},
			Fields:         fields,
			ChecksumsValid: validateTD3(b),
			Format:         "TD3",
		}, true
	}
	return MRZ{}, false
}

// parseMRZPair 解析已归一化的 TD3 两行。
func parseMRZPair(a, b string) Result {
	namePart := a[5:]
	parts := strings.SplitN(namePart, "<<", 2)
	surname := strings.ReplaceAll(parts[0], "<", " ")
	given := ""
	if len(parts) == 2 {
		given = strings.ReplaceAll(parts[1], "<", " ")
	}
	fullName := strings.TrimSpace(strings.TrimSpace(given) + " " + strings.TrimSpace(surname))
	return Result{
		FullName:       fullName,
		DateOfBirth:    mrzDate(b[13:19], true),
		IDNumber:       strings.ReplaceAll(b[0:9], "<", ""),
		ExpiryDate:     mrzDate(b[21:27], false),
		IssuingCountry: alpha3ToAlpha2(a[2:5]),
	}
}

// validateTD3 校验 TD3 的证件号、出生、到期、个人号与复合校验位。
func validateTD3(line string) bool {
	if len(line) < 44 {
		return false
	}
	checks := []struct {
		start, end, pos int
	}{
		{0, 9, 9},
		{13, 19, 19},
		{21, 27, 27},
		{28, 42, 42},
	}
	for _, check := range checks {
		if mrzCheckDigit(line[check.start:check.end]) != digit(line[check.pos]) {
			return false
		}
	}
	composite := line[0:10] + line[13:20] + line[21:28] + line[28:43]
	return mrzCheckDigit(composite) == digit(line[43])
}

// mrzCheckDigit 计算 MRZ 7-3-1 加权校验位。
func mrzCheckDigit(value string) int {
	sum := 0
	for i, r := range []byte(value) {
		sum += mrzCharValue(r) * [3]int{7, 3, 1}[i%3]
	}
	return sum % 10
}

// mrzCharValue 将数字、字母和填充符映射为 MRZ 值。
func mrzCharValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'Z':
		return int(c-'A') + 10
	default:
		return 0
	}
}

// digit 把 ASCII 数字转为整数，非数字返回 -1。
func digit(c byte) int {
	if c < '0' || c > '9' {
		return -1
	}
	return int(c - '0')
}

// alpha3ToAlpha2 覆盖巴西护照及常见 MRZ 签发国。
func alpha3ToAlpha2(code string) string {
	switch strings.ToUpper(code) {
	case "BRA":
		return "BR"
	case "USA":
		return "US"
	case "GBR":
		return "GB"
	case "CHN":
		return "CN"
	case "CAN":
		return "CA"
	default:
		return ""
	}
}
