// Package barcode 提供纯 Go 的二维码、PDF417 与一维条码解码。
package barcode

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	zxinggo "github.com/ericlevine/zxinggo"
	_ "github.com/ericlevine/zxinggo/aztec"
	"github.com/ericlevine/zxinggo/binarizer"
	_ "github.com/ericlevine/zxinggo/datamatrix"
	_ "github.com/ericlevine/zxinggo/oned"
	_ "github.com/ericlevine/zxinggo/pdf417"
	_ "github.com/ericlevine/zxinggo/qrcode"
)

// Result 是条码解码结果。
type Result struct {
	Format string
	Text   string
}

// Decode 自动识别常见二维码、PDF417 与一维条码。
func Decode(data []byte) (Result, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Result{}, err
	}
	source := zxinggo.NewImageLuminanceSource(img)
	bitmap := zxinggo.NewBinaryBitmap(binarizer.NewHybrid(source))
	result, err := zxinggo.Decode(bitmap, nil)
	if err != nil || result == nil {
		return Result{}, fmt.Errorf("未识别到条码")
	}
	return Result{Format: result.Format.String(), Text: result.Text}, nil
}
