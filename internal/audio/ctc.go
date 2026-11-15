package audio

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadTokens 读取 NeMo tokens.txt，返回按 id 索引的 token 表。
func LoadTokens(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tokens []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("tokens: invalid line %q", line)
		}
		id, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			return nil, fmt.Errorf("tokens: invalid id in %q: %w", line, err)
		}
		token := strings.Join(fields[:len(fields)-1], " ")
		for len(tokens) <= id {
			tokens = append(tokens, "")
		}
		tokens[id] = token
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

// GreedyDecode 对 CTC logprobs 做 best-path 解码：逐帧 argmax、折叠重复、去 blank。
func GreedyDecode(logprobs [][]float32, tokens []string) (string, []int) {
	var picked []string
	var ids []int
	last := -1
	for _, row := range logprobs {
		best := 0
		for i := 1; i < len(row); i++ {
			if row[i] > row[best] {
				best = i
			}
		}
		if best == last {
			continue
		}
		last = best
		if best == 0 {
			continue
		}
		if best < len(tokens) && tokens[best] != "" {
			picked = append(picked, tokens[best])
			ids = append(ids, best)
		}
	}
	return Detokenize(picked), ids
}

// Detokenize 将 NeMo BPE token 序列拼成可读文本。
func Detokenize(tokens []string) string {
	var b strings.Builder
	for _, token := range tokens {
		if token == "" || token == "<unk>" || token == "<blk>" {
			continue
		}
		if strings.HasPrefix(token, "▁") {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			token = strings.TrimPrefix(token, "▁")
		}
		b.WriteString(token)
	}
	return strings.TrimSpace(b.String())
}

var digitWords = map[string]byte{
	"zero": '0', "one": '1', "two": '2', "three": '3', "four": '4',
	"five": '5', "six": '6', "seven": '7', "eight": '8', "nine": '9',
}

// DigitsOnly 从转录文本中提取数字口令，支持 0-9 与英文数字词。
func DigitsOnly(text string) string {
	var out []byte
	for _, field := range strings.Fields(strings.ToLower(text)) {
		field = strings.Trim(field, ".,!?;:\"'")
		if len(field) == 1 && field[0] >= '0' && field[0] <= '9' {
			out = append(out, field[0])
			continue
		}
		if d, ok := digitWords[field]; ok {
			out = append(out, d)
		}
	}
	return string(out)
}
