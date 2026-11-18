package ocr

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadDict 验证首行 blank、字符映射与尾随换行产生的末位 blank。
func TestLoadDict(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dict.txt")
	if err := os.WriteFile(path, []byte("\nA\nB\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dict, err := LoadDict(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(dict) != 4 || dict[0] != 0 || dict[1] != 'A' || dict[2] != 'B' || dict[3] != 0 {
		t.Fatalf("unexpected dict: %v", dict)
	}

	noTrailing := filepath.Join(dir, "dict2.txt")
	if err := os.WriteFile(noTrailing, []byte("\nA\nB"), 0o644); err != nil {
		t.Fatal(err)
	}
	dict2, err := LoadDict(noTrailing)
	if err != nil {
		t.Fatal(err)
	}
	if len(dict2) != 3 || dict2[2] != 'B' {
		t.Fatalf("unexpected dict without trailing newline: %v", dict2)
	}
}

// TestOrientationAngles 验证方向配置展开。
func TestOrientationAngles(t *testing.T) {
	if got := orientationAngles("auto"); len(got) != 4 || got[0] != 0 || got[3] != 270 {
		t.Fatalf("auto angles: %v", got)
	}
	if got := orientationAngles("90"); len(got) != 1 || got[0] != 90 {
		t.Fatalf("90 angles: %v", got)
	}
	if got := orientationAngles(""); len(got) != 1 || got[0] != 0 {
		t.Fatalf("default angles: %v", got)
	}
}
