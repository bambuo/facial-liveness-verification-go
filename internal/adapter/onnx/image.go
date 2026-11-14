// Package onnx 实现纯 Go ONNX 推理与图像前后处理。
package onnx

import (
	"bytes"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"

	xdraw "golang.org/x/image/draw"

	"facial-liveness-verification-go/internal/domain"
)

// referenceLandmarks 是 ArcFace 112×112 标准 5 点。
var referenceLandmarks = [5][2]float64{
	{38.2946, 51.6963},
	{73.5318, 51.5014},
	{56.0252, 71.7366},
	{41.5493, 92.3655},
	{70.7299, 92.2041},
}

// decodeImage 解码 JPEG/PNG 并统一为 RGBA。
func decodeImage(data []byte) (*image.RGBA, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, domain.New(domain.CodeLowQuality, "图片解码失败（仅支持 JPEG/PNG）")
	}
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	return out, nil
}

// letterbox 等比缩放并贴左上角，右侧与下方补 0。
func letterbox(img *image.RGBA, size int) *image.RGBA {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	scale := math.Min(float64(size)/float64(w), float64(size)/float64(h))
	paddedW := maxInt(1, int(math.Round(float64(w)*scale)))
	paddedH := maxInt(1, int(math.Round(float64(h)*scale)))
	canvas := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.BiLinear.Scale(canvas, image.Rect(0, 0, paddedW, paddedH), img, img.Bounds(), draw.Src, nil)
	return canvas
}

// toCHW 将 RGBA 转为 CHW float，bgr 控制前两通道交换。
func toCHW(img *image.RGBA, bgr bool) []float32 {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	plane := w * h
	out := make([]float32, 3*plane)
	pix := img.Pix
	stride := img.Stride
	c0, c2 := 0, 2
	if bgr {
		c0, c2 = 2, 0
	}
	for y := 0; y < h; y++ {
		row := y * stride
		for x := 0; x < w; x++ {
			i := row + x*4
			p := y*w + x
			out[p] = float32(pix[i+c0])
			out[plane+p] = float32(pix[i+1])
			out[2*plane+p] = float32(pix[i+c2])
		}
	}
	return out
}

// cropWithMargin 以人脸框中心做方形裁剪并缩放到 inputSize。
func cropWithMargin(img *image.RGBA, b domain.FaceBox, cropScale float64, inputSize int) (*image.RGBA, error) {
	cx := (b.X1 + b.X2) / 2
	cy := (b.Y1 + b.Y2) / 2
	halfW := (b.X2 - b.X1) / 2
	halfH := (b.Y2 - b.Y1) / 2
	side := math.Max(halfW, halfH) * cropScale
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	x1 := int(math.Max(0, math.Floor(cx-side)))
	y1 := int(math.Max(0, math.Floor(cy-side)))
	x2 := int(math.Min(float64(w), math.Floor(cx+side)))
	y2 := int(math.Min(float64(h), math.Floor(cy+side)))
	if x2-x1 <= 0 || y2-y1 <= 0 {
		return nil, domain.New(domain.CodeLivenessFailed, "裁剪入参为空")
	}
	out := image.NewRGBA(image.Rect(0, 0, inputSize, inputSize))
	xdraw.BiLinear.Scale(out, out.Bounds(), img, image.Rect(x1, y1, x2, y2), draw.Src, nil)
	return out, nil
}

// alignCrop 按 5 点相似变换对齐并缩放到 112×112。
func alignCrop(img *image.RGBA, landmarks [5][2]float64) *image.RGBA {
	var cpx, cpy, cax, cay float64
	for i := 0; i < 5; i++ {
		cpx += landmarks[i][0]
		cpy += landmarks[i][1]
		cax += referenceLandmarks[i][0]
		cay += referenceLandmarks[i][1]
	}
	cpx /= 5
	cpy /= 5
	cax /= 5
	cay /= 5

	var a, b, norm float64
	for i := 0; i < 5; i++ {
		px := landmarks[i][0] - cpx
		py := landmarks[i][1] - cpy
		ax := referenceLandmarks[i][0] - cax
		ay := referenceLandmarks[i][1] - cay
		a += px*ax + py*ay
		b += py*ax - px*ay
		norm += px*px + py*py
	}
	theta := math.Atan2(b, a)
	scale := math.Sqrt(a*a+b*b) / norm
	cos, sin := math.Cos(theta), math.Sin(theta)
	tx := cax - scale*(cos*cpx-sin*cpy)
	ty := cay - scale*(sin*cpx+cos*cpy)

	m00, m01 := scale*cos, -scale*sin
	m10, m11 := scale*sin, scale*cos
	det := m00*m11 - m01*m10
	inv00, inv01 := m11/det, -m01/det
	inv10, inv11 := -m10/det, m00/det

	out := image.NewRGBA(image.Rect(0, 0, 112, 112))
	for y := 0; y < 112; y++ {
		for x := 0; x < 112; x++ {
			dx := float64(x) - tx
			dy := float64(y) - ty
			sx := inv00*dx + inv01*dy
			sy := inv10*dx + inv11*dy
			r, g, bl := sampleBilinear(img, sx, sy)
			i := out.PixOffset(x, y)
			out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = r, g, bl, 255
		}
	}
	return out
}

// sampleBilinear 双线性采样，越界返回 0。
func sampleBilinear(img *image.RGBA, x, y float64) (byte, byte, byte) {
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if x0 < 0 || y0 < 0 || x0+1 >= w || y0+1 >= h {
		return 0, 0, 0
	}
	fx, fy := x-float64(x0), y-float64(y0)
	p00, p10 := img.RGBAAt(x0, y0), img.RGBAAt(x0+1, y0)
	p01, p11 := img.RGBAAt(x0, y0+1), img.RGBAAt(x0+1, y0+1)
	lerp := func(a, b, t float64) float64 { return a + (b-a)*t }
	return byte(lerp(lerp(float64(p00.R), float64(p10.R), fx), lerp(float64(p01.R), float64(p11.R), fx), fy)),
		byte(lerp(lerp(float64(p00.G), float64(p10.G), fx), lerp(float64(p01.G), float64(p11.G), fx), fy)),
		byte(lerp(lerp(float64(p00.B), float64(p10.B), fx), lerp(float64(p01.B), float64(p11.B), fx), fy))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func largestFace(boxes []domain.FaceBox) *domain.FaceBox {
	var best *domain.FaceBox
	bestArea := -1.0
	for i := range boxes {
		b := &boxes[i]
		area := math.Max(0, b.X2-b.X1) * math.Max(0, b.Y2-b.Y1)
		if area > bestArea {
			bestArea = area
			best = b
		}
	}
	return best
}
