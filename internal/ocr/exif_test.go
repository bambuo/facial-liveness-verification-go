package ocr

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"testing"
)

// appendExifOrientation 往 JPEG 的 SOI 之后插入一个带 Orientation 的最小 Exif APP1 段。
func appendExifOrientation(t *testing.T, src []byte, orientation uint16) []byte {
	t.Helper()
	tiff := []byte("II\x2a\x00\x08\x00\x00\x00\x01\x00")
	entry := make([]byte, 12)
	binary.LittleEndian.PutUint16(entry[0:2], 0x0112)
	binary.LittleEndian.PutUint16(entry[2:4], 3)
	binary.LittleEndian.PutUint32(entry[4:8], 1)
	binary.LittleEndian.PutUint16(entry[8:10], orientation)
	tiff = append(tiff, entry...)
	tiff = append(tiff, 0x00, 0x00, 0x00, 0x00)
	body := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1, byte((len(body) + 2) >> 8), byte((len(body) + 2) & 0xFF)}
	seg = append(seg, body...)
	out := append([]byte{src[0], src[1]}, seg...)
	return append(out, src[2:]...)
}

// TestExifOrientation 覆盖解析器对合法、非法与非 JPEG 输入的行为。
func TestExifOrientation(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	plain := buf.Bytes()

	if got := exifOrientation(plain); got != 0 {
		t.Fatalf("无 EXIF 应为 0，得到 %d", got)
	}
	for _, tc := range []struct {
		orientation uint16
		want        int
		angle       int
	}{
		{1, 1, 0},
		{3, 3, 180},
		{6, 6, 90},
		{8, 8, 270},
		{2, 2, 0},
		{9, 0, 0},
	} {
		data := appendExifOrientation(t, plain, tc.orientation)
		if got := exifOrientation(data); got != tc.want {
			t.Fatalf("orientation=%d: got %d want %d", tc.orientation, got, tc.want)
		}
		if got := exifRotationAngle(tc.want); got != tc.angle {
			t.Fatalf("angle(orientation=%d): got %d want %d", tc.want, got, tc.angle)
		}
	}
	if got := exifOrientation([]byte{0x89, 0x50}); got != 0 {
		t.Fatalf("PNG 应为 0，得到 %d", got)
	}
	if got := exifRotationAngle(5); got != 0 {
		t.Fatalf("镜像方向 5 应为 0，得到 %d", got)
	}
}

// TestEffectiveOrientation 验证 EXIF 存在时 auto 折叠为单方向，其余配置保持原语义。
func TestEffectiveOrientation(t *testing.T) {
	for _, tc := range []struct {
		requested, cfg string
		hasExif        bool
		want           string
	}{
		{"", "auto", true, "0"},
		{"", "auto", false, "auto"},
		{"", "0", true, "0"},
		{"90", "auto", true, "90"},
		{"270", "0", false, "270"},
	} {
		if got := effectiveOrientation(tc.requested, tc.cfg, tc.hasExif); got != tc.want {
			t.Fatalf("effectiveOrientation(%q,%q,%v)=%q want %q", tc.requested, tc.cfg, tc.hasExif, got, tc.want)
		}
	}
}
