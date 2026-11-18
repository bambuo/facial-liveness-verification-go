// 生成用于 OCR 回归测试的文本图片。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// main 渲染指定文本为 PNG，供 OCR 回归与多语言测试。
func main() {
	out := flag.String("out", "/tmp/ocr-text.png", "输出 PNG")
	text := flag.String("text", "姓名：张三 出生日期：1990-01-01 证件号码：1234567890", "文本")
	fontPath := flag.String("font", "/System/Library/Fonts/Supplemental/Arial Unicode.ttf", "字体文件")
	size := flag.Float64("size", 48, "字号")
	padding := flag.Int("padding", 32, "边距")
	flag.Parse()

	buf, err := os.ReadFile(*fontPath)
	if err != nil {
		fatal(err)
	}
	f, err := opentype.Parse(buf)
	if err != nil {
		fatal(err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: *size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		fatal(err)
	}
	width := font.MeasureString(face, *text).Ceil() + 2**padding
	height := int(*size*1.4) + 2**padding
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.P(*padding, *padding+face.Metrics().Ascent.Ceil()),
	}
	drawer.DrawString(*text)
	fh, err := os.Create(*out)
	if err != nil {
		fatal(err)
	}
	defer fh.Close()
	if err := png.Encode(fh, img); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s %dx%d\n", *out, width, height)
}

// fatal 打印错误并以非 0 退出。
func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
