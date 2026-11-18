// Package fields 提供证件字段抽取规则（v1）。
package fields

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Result 是证件正面/护照的字段抽取结果。
type Result struct {
	FullName       string
	DateOfBirth    string
	IDNumber       string
	ExpiryDate     string
	IssuingCountry string
}

// Extract 按 MRZ 优先、国家规则其次、通用规则兜底抽取字段。
func Extract(text, issuingCountry string) Result {
	r := Result{IssuingCountry: strings.ToUpper(strings.TrimSpace(issuingCountry))}
	if mrz, ok := parseMRZ(text); ok {
		r.FullName = mrz.FullName
		r.DateOfBirth = mrz.DateOfBirth
		r.IDNumber = mrz.IDNumber
		r.ExpiryDate = mrz.ExpiryDate
	}
	switch r.IssuingCountry {
	case "US", "USA", "CA", "CAN":
		extractUS(text, &r)
	case "GB", "UK", "GBR":
		extractUK(text, &r)
	case "CN", "CHN":
		extractCN(text, &r)
	case "BR", "BRA":
		extractBR(text, &r)
	}
	extractGeneric(text, &r)
	return r
}

type mrzFields struct {
	FullName, DateOfBirth, IDNumber, ExpiryDate string
}

// mrzCandidate 归一化一行 MRZ：去空格、截断/补尾 < 到 44 位。
func mrzCandidate(line string, first bool) string {
	compact := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(line)), " ", "")
	if idx := strings.Index(compact, "P<"); idx > 0 && idx < 8 {
		compact = compact[idx:]
	}
	if len(compact) >= 44 {
		compact = compact[:44]
	} else if len(compact) >= 40 && first {
		compact += strings.Repeat("<", 44-len(compact))
	} else if len(compact) >= 40 {
		compact += strings.Repeat("<", 44-len(compact))
	} else {
		return ""
	}
	if first {
		switch compact[0] {
		case 'P', 'V', 'I', 'A', 'C':
		default:
			return ""
		}
	}
	return compact
}

// parseMRZ 解析 TD3 两行 MRZ（护照）。
func parseMRZ(text string) (mrzFields, bool) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := 0; i+1 < len(lines); i++ {
		a := mrzCandidate(lines[i], true)
		b := mrzCandidate(lines[i+1], false)
		if a == "" || b == "" {
			continue
		}
		namePart := a[5:]
		parts := strings.SplitN(namePart, "<<", 2)
		surname := strings.ReplaceAll(parts[0], "<", " ")
		given := ""
		if len(parts) == 2 {
			given = strings.ReplaceAll(parts[1], "<", " ")
		}
		fullName := strings.TrimSpace(strings.TrimSpace(given) + " " + strings.TrimSpace(surname))
		id := strings.ReplaceAll(b[0:9], "<", "")
		return mrzFields{
			FullName:    fullName,
			DateOfBirth: mrzDate(b[13:19], true),
			IDNumber:    id,
			ExpiryDate:  mrzDate(b[21:27], false),
		}, true
	}
	return mrzFields{}, false
}

// mrzDate 解析 YYMMDD；出生日期按当前年份推断世纪，到期日固定 20xx。
func mrzDate(raw string, birth bool) string {
	if len(raw) != 6 {
		return ""
	}
	yy, err1 := strconv.Atoi(raw[0:2])
	mm, err2 := strconv.Atoi(raw[2:4])
	dd, err3 := strconv.Atoi(raw[4:6])
	if err1 != nil || err2 != nil || err3 != nil {
		return ""
	}
	year := 2000 + yy
	if birth && year > time.Now().Year() {
		year -= 100
	}
	return formatDate(year, mm, dd)
}

var (
	usName = regexp.MustCompile(`(?i)\b(?:FULL NAME|NAME|SURNAME)\b\s*[:#]?\s*(.+)`)
	usDOB  = regexp.MustCompile(`(?i)\b(?:DOB|DATE OF BIRTH|BIRTH)\b\s*[:#]?\s*(.+)`)
	usExp  = regexp.MustCompile(`(?i)\b(?:EXP|EXPIRY|EXPIRES|VALID UNTIL)\b\s*[:#]?\s*(.+)`)
	usID   = regexp.MustCompile(`(?i)\b(?:DL|DLN|LIC(?:ENSE)?|OL|ID)\b\s*(?:NO\.?|#)?\s*[:#]?\s*([A-Z]{0,3}\d[\dA-Z-]{4,16})`)
	ukID   = regexp.MustCompile(`\b[A-Z9]{5}\d{6}[A-Z9]{2}\d[A-Z]{2}\b`)
	cnID   = regexp.MustCompile(`\b\d{17}[\dXx]\b`)
)

// extractUS 处理美国/加拿大驾照与身份证常见标签。
func extractUS(text string, r *Result) {
	if r.FullName == "" {
		r.FullName = cleanName(findLabel(usName, text))
	}
	if r.DateOfBirth == "" {
		r.DateOfBirth = parseDate(findLabel(usDOB, text), "mdy")
	}
	if r.ExpiryDate == "" {
		r.ExpiryDate = parseDate(findLabel(usExp, text), "mdy")
	}
	if r.IDNumber == "" {
		if m := usID.FindStringSubmatch(text); len(m) == 2 {
			r.IDNumber = strings.ToUpper(strings.Trim(m[1], "-"))
		}
	}
}

// extractUK 处理英国驾照：16 位证号与 DD/MM/YYYY 日期。
func extractUK(text string, r *Result) {
	if r.FullName == "" {
		r.FullName = cleanName(findLabel(usName, text))
	}
	if r.DateOfBirth == "" {
		r.DateOfBirth = parseDate(findLabel(usDOB, text), "dmy")
	}
	if r.ExpiryDate == "" {
		r.ExpiryDate = parseDate(findLabel(usExp, text), "dmy")
	}
	if r.IDNumber == "" {
		r.IDNumber = ukID.FindString(text)
	}
}

var (
	cnName = regexp.MustCompile(`姓名\s*[:：]?\s*(.+)`)
	cnDOB  = regexp.MustCompile(`(?:出生日期|出生)\s*[:：]?\s*(.+)`)
	cnExp  = regexp.MustCompile(`(?:有效期限|有效期至|到期)\s*[:：]?\s*(.+)`)
)

// extractCN 处理中国身份证：18 位号码内含出生日期。
func extractCN(text string, r *Result) {
	if r.FullName == "" {
		r.FullName = cleanName(findLabel(cnName, text))
	}
	if m := cnID.FindString(text); m != "" {
		if r.IDNumber == "" {
			r.IDNumber = strings.ToUpper(m)
		}
		if r.DateOfBirth == "" {
			if y, _ := strconv.Atoi(m[6:10]); y > 1900 {
				mo, _ := strconv.Atoi(m[10:12])
				d, _ := strconv.Atoi(m[12:14])
				r.DateOfBirth = formatDate(y, mo, d)
			}
		}
	}
	if r.DateOfBirth == "" {
		r.DateOfBirth = parseDate(findLabel(cnDOB, text), "ymd")
	}
	if r.ExpiryDate == "" {
		r.ExpiryDate = parseDate(findLabel(cnExp, text), "ymd")
	}
}

var brLabels = struct {
	surname  []string
	given    []string
	dob      []string
	expiry   []string
	passport []string
}{
	surname:  []string{"SOBRENOME / SURNAME", "SURNAME / SOBRENOME", "SOBRENOME"},
	given:    []string{"NOME / GIVEN NAMES", "GIVEN NAMES", "NOME"},
	dob:      []string{"DATA DE NASCIMENTO", "DATE OF BIRTH"},
	expiry:   []string{"VALIDO ATE", "VALIDO ATÉ", "VÁLIDO ATE", "VÁLIDO ATÉ", "DATE OF EXPIRY", "VALID UNTIL"},
	passport: []string{"PASSAPORTE N", "PASSPORT NO", "PASSPORT N"},
}

// extractBR 使用巴西护照可视标签补齐/校正 MRZ 字段。
func extractBR(text string, r *Result) {
	surname := labelNextLine(text, brLabels.surname, []string{"GIVEN NAMES"})
	given := labelNextLine(text, brLabels.given, []string{"SOBRENOME"})
	if surname != "" && given != "" {
		r.FullName = strings.TrimSpace(given + " " + surname)
	}
	if r.DateOfBirth == "" {
		if d := parseDate(labelNextLine(text, brLabels.dob, nil), "dmy"); d != "" {
			r.DateOfBirth = d
		}
	}
	if r.ExpiryDate == "" {
		if d := parseDate(labelNextLine(text, brLabels.expiry, nil), "dmy"); d != "" {
			r.ExpiryDate = d
		}
	}
	if r.IDNumber == "" {
		if id := labelNextLine(text, brLabels.passport, nil); id != "" {
			best := ""
			for _, token := range strings.Fields(id) {
				if len(token) >= 6 && len(token) > len(best) {
					best = token
				}
			}
			r.IDNumber = best
		}
	}
}

// labelNextLine 返回标签同行的剩余值或下一行值。
func labelNextLine(text string, labels, exclude []string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		up := strings.ToUpper(strings.TrimSpace(line))
		if up == "" {
			continue
		}
		skip := false
		for _, ex := range exclude {
			if strings.Contains(up, strings.ToUpper(ex)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		for _, label := range labels {
			l := strings.ToUpper(label)
			if !strings.Contains(up, l) {
				continue
			}
			if idx := strings.Index(up, l); idx >= 0 {
				rest := strings.TrimSpace(up[idx+len(l):])
				rest = strings.TrimLeft(rest, " /:：#-")
				if rest != "" && !looksLikeLabel(rest) {
					return strings.TrimSpace(rest)
				}
			}
			for j := i + 1; j < len(lines) && j <= i+2; j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" || looksLikeLabel(next) {
					continue
				}
				return next
			}
		}
	}
	return ""
}

func looksLikeLabel(s string) bool {
	u := strings.ToUpper(strings.TrimSpace(s))
	for _, l := range []string{"SOBRENOME", "SURNAME", "GIVEN NAMES", "DATE OF BIRTH", "DATA DE NASCIMENTO", "DATE OF EXPIRY", "VALIDO", "VÁLIDO", "PASSAPORTE", "PASSPORT", "NACIONALIDADE", "SEXO", "LOCAL"} {
		if strings.Contains(u, l) {
			return true
		}
	}
	return false
}

var (
	genericName = regexp.MustCompile(`(?i)\b(?:NAME|NAMA|NOMBRE|NOME|姓名|氏名)\b\s*[:：]?\s*(.+)`)
	genericDOB  = regexp.MustCompile(`(?i)\b(?:DOB|BIRTH|出生|生日|DATE OF BIRTH)\b\s*[:：]?\s*(.+)`)
	genericExp  = regexp.MustCompile(`(?i)\b(?:EXP|EXPIRY|EXPIRES|VALID|有效|到期)\b\s*[:：]?\s*(.+)`)
	anyDate     = regexp.MustCompile(`\b\d{4}[-/.]\d{1,2}[-/.]\d{1,2}\b|\b\d{1,2}[-/.]\d{1,2}[-/.]\d{4}\b`)
	genericID   = regexp.MustCompile(`\b[A-Z0-9]{6,20}\b`)
)

// extractGeneric 作为兜底：标签优先，日期与编号用通用正则。
func extractGeneric(text string, r *Result) {
	if r.FullName == "" {
		r.FullName = cleanName(findLabel(genericName, text))
	}
	if r.DateOfBirth == "" {
		r.DateOfBirth = parseDate(findLabel(genericDOB, text), "ymd")
	}
	if r.ExpiryDate == "" {
		r.ExpiryDate = parseDate(findLabel(genericExp, text), "ymd")
	}
	dates := allDates(text)
	if r.DateOfBirth == "" && len(dates) > 0 {
		r.DateOfBirth = dates[0]
	}
	if r.ExpiryDate == "" && len(dates) > 1 {
		r.ExpiryDate = dates[len(dates)-1]
	}
	if r.IDNumber == "" {
		r.IDNumber = pickID(text)
	}
	if r.FullName == "" {
		r.FullName = firstPersonName(text)
	}
}

func findLabel(re *regexp.Regexp, text string) string {
	for _, line := range strings.Split(text, "\n") {
		if m := re.FindStringSubmatch(line); len(m) == 2 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func cleanName(raw string) string {
	name := strings.TrimSpace(raw)
	name = strings.Trim(name, "\"'")
	if i := strings.IndexAny(name, "|,;"); i > 0 {
		name = name[:i]
	}
	return strings.TrimSpace(name)
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "FEV": 2, "MAR": 3, "APR": 4, "ABR": 4, "MAY": 5, "MAI": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "AGO": 8, "SEP": 9, "SET": 9, "OCT": 10, "OUT": 10, "NOV": 11, "DEC": 12, "DEZ": 12,
}

// parseDate 解析 YYYY-MM-DD、MM/DD/YYYY、DD/MM/YYYY 与 DD MMM YYYY。
func parseDate(raw, order string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	monthName := regexp.MustCompile(`(?i)(\d{1,2})\s+([A-Z]{3,9})(?:\s*/\s*[A-Z]{3,9})?\s+(\d{4})`).FindStringSubmatch(raw)
	if len(monthName) == 4 {
		d, _ := strconv.Atoi(monthName[1])
		m := monthNames[strings.ToUpper(monthName[2][:3])]
		y, _ := strconv.Atoi(monthName[3])
		return formatDate(y, m, d)
	}
	cn := regexp.MustCompile(`(\d{4})\s*年\s*(\d{1,2})\s*月\s*(\d{1,2})\s*日`).FindStringSubmatch(raw)
	if len(cn) == 4 {
		y, _ := strconv.Atoi(cn[1])
		m, _ := strconv.Atoi(cn[2])
		d, _ := strconv.Atoi(cn[3])
		return formatDate(y, m, d)
	}
	nums := regexp.MustCompile(`\d+`).FindAllString(raw, -1)
	if len(nums) != 3 {
		return ""
	}
	a, _ := strconv.Atoi(nums[0])
	b, _ := strconv.Atoi(nums[1])
	c, _ := strconv.Atoi(nums[2])
	switch {
	case len(nums[0]) == 4:
		return formatDate(a, b, c)
	case order == "dmy":
		return formatDate(c, b, a)
	case order == "mdy":
		return formatDate(c, a, b)
	default:
		if c > 31 {
			return formatDate(c, b, a)
		}
		return formatDate(c, a, b)
	}
}

func formatDate(year, month, day int) string {
	if year < 1900 || year > 2200 || month < 1 || month > 12 || day < 1 || day > 31 {
		return ""
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func allDates(text string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, raw := range anyDate.FindAllString(text, -1) {
		d := parseDate(raw, "ymd")
		if d != "" && !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

func pickID(text string) string {
	for _, token := range genericID.FindAllString(text, -1) {
		digits := 0
		for _, r := range token {
			if r >= '0' && r <= '9' {
				digits++
			}
		}
		if digits >= 4 && !anyDate.MatchString(token) {
			return token
		}
	}
	return ""
}

func firstPersonName(text string) string {
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || len(fields) > 5 {
			continue
		}
		ok := true
		for _, f := range fields {
			if len([]rune(f)) < 2 || strings.ContainsAny(f, "0123456789:/") {
				ok = false
				break
			}
			r := []rune(f)
			if r[0] < 'A' || r[0] > 'Z' {
				ok = false
				break
			}
		}
		if ok {
			return strings.Join(fields, " ")
		}
	}
	return ""
}
