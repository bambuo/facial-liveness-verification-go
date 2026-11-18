// Package tamper 是参考实现 SharpTamperDetector + FrequencyAnalyzer 的 Go 移植。
package tamper

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"
	"math"
	"math/cmplx"
)

// Result 是篡改检测结果，score 越大越可信。
type Result struct {
	Score       float64
	Flags       []string
	IsAuthentic bool
	Details     map[string]any
}

// Analyze 执行 ELA、熵、EXIF、颜色、双压缩与 FFT 六项检测。
func Analyze(data []byte) (Result, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return Result{}, err
	}
	flags := make([]string, 0)
	score := 1.0
	details := map[string]any{}

	stats := imageStats(img)
	reencoded, err := reencode(data, 85)
	if err == nil {
		reStats := imageStats(reencoded)
		elaDiff := math.Abs(stats.mean - reStats.mean)
		details["ela"] = map[string]any{"diff": elaDiff, "flagged": elaDiff > 15}
		if elaDiff > 15 {
			flags = append(flags, "HIGH_ELA_DIFFERENCE")
			score -= 0.25
		}
	}
	entropy := stats.stdev
	details["entropy"] = map[string]any{"value": entropy, "flagged": entropy < 5}
	if entropy < 5 {
		flags = append(flags, "LOW_ENTROPY_REGIONS")
		score -= 0.15
	}
	hasExif := bytes.Contains(data, []byte("Exif\x00\x00"))
	exifFlagged := !hasExif && format == "jpeg"
	details["exif"] = map[string]any{"present": hasExif, "flagged": exifFlagged}
	if exifFlagged {
		flags = append(flags, "MISSING_EXIF_JPEG")
		score -= 0.08
	}
	colorAnomalies := colorAnomalies(stats)
	details["colorAnomaly"] = colorAnomalies
	if len(colorAnomalies) > 0 {
		flags = append(flags, colorAnomalies...)
		score -= 0.15
	}
	if double, variance := doubleCompression(data, img); double {
		details["doubleCompression"] = map[string]any{"detected": true, "regionVariance": variance}
		flags = append(flags, "DOUBLE_COMPRESSION_DETECTED")
		score -= 0.20
	} else {
		details["doubleCompression"] = map[string]any{"detected": false, "regionVariance": variance}
	}
	if freq, err := frequencyAnalysis(img); err == nil {
		details["frequency"] = freq
		for _, anomaly := range freq.Anomalies {
			if anomaly != "ANALYSIS_FAILED" {
				flags = append(flags, "GAN_SPECTRAL_ANOMALY")
				score -= 0.25
				break
			}
		}
	}
	if score < 0 {
		score = 0
	}
	return Result{Score: score, Flags: flags, IsAuthentic: score >= 0.7, Details: details}, nil
}

type stats struct {
	mean, stdev float64
	channels    [][2]float64
}

func imageStats(img image.Image) stats {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width == 0 || height == 0 {
		return stats{}
	}
	var sums, squares [3]float64
	count := float64(width * height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			values := [3]float64{float64(r) * 255 / 65535, float64(g) * 255 / 65535, float64(b) * 255 / 65535}
			for i, value := range values {
				sums[i] += value
				squares[i] += value * value
			}
		}
	}
	var totalMean, totalStdev float64
	channels := make([][2]float64, 3)
	for i := range channels {
		mean := sums[i] / count
		variance := math.Max(0, squares[i]/count-mean*mean)
		stdev := math.Sqrt(variance)
		channels[i] = [2]float64{mean, stdev}
		totalMean += mean
		totalStdev += stdev
	}
	return stats{mean: totalMean / 3, stdev: totalStdev / 3, channels: channels}
}

func colorAnomalies(s stats) []string {
	anomalies := make([]string, 0)
	avgStdev := (s.channels[0][1] + s.channels[1][1] + s.channels[2][1]) / 3
	if avgStdev < 20 {
		anomalies = append(anomalies, "COLOR_ABNORMAL_UNIFORMITY")
	}
	maxStdev := math.Max(s.channels[0][1], math.Max(s.channels[1][1], s.channels[2][1]))
	minStdev := math.Min(s.channels[0][1], math.Min(s.channels[1][1], s.channels[2][1]))
	if minStdev > 0 && maxStdev/minStdev > 4 {
		anomalies = append(anomalies, "COLOR_CHANNEL_DECORRELATION")
	}
	meanRange := math.Max(s.channels[0][0], math.Max(s.channels[1][0], s.channels[2][0])) -
		math.Min(s.channels[0][0], math.Min(s.channels[1][0], s.channels[2][0]))
	if meanRange < 3 && avgStdev > 30 {
		anomalies = append(anomalies, "COLOR_DESATURATED_RGB")
	}
	return anomalies
}

func reencode(data []byte, quality int) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	out, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	return out, err
}

func doubleCompression(data []byte, original image.Image) (bool, float64) {
	bounds := original.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < 128 || height < 128 {
		return false, 0
	}
	img85, err85 := reencode(data, 85)
	img50, err50 := reencode(data, 50)
	if err85 != nil || err50 != nil {
		return false, 0
	}
	diffs := make([]float64, 0, 16)
	for gy := 0; gy < 4; gy++ {
		for gx := 0; gx < 4; gx++ {
			rect := image.Rect(gx*width/4, gy*height/4, (gx+1)*width/4, (gy+1)*height/4)
			origMean := regionMean(original, rect)
			diff85 := math.Abs(origMean - regionMean(img85, rect))
			diff50 := math.Abs(origMean - regionMean(img50, rect))
			if diff50 > 0 {
				diffs = append(diffs, diff85/diff50)
			}
		}
	}
	if len(diffs) < 4 {
		return false, 0
	}
	mean := 0.0
	for _, value := range diffs {
		mean += value
	}
	mean /= float64(len(diffs))
	variance := 0.0
	for _, value := range diffs {
		variance += (value - mean) * (value - mean)
	}
	variance /= float64(len(diffs))
	return variance > 0.15, variance
}

func regionMean(img image.Image, rect image.Rectangle) float64 {
	bounds := img.Bounds()
	rect = rect.Intersect(bounds)
	if rect.Empty() {
		return 0
	}
	var sum float64
	count := 0
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			sum += float64(r+g+b) * 255 / (3 * 65535)
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

type frequencyResult struct {
	GanScore    float64
	EnergyRatio float64
	Anomalies   []string
}

func frequencyAnalysis(img image.Image) (frequencyResult, error) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	n := min(nextPow2(min(width, 1024)), nextPow2(min(height, 1024)))
	gray := make([]float64, n*n)
	startX := max(0, (width-n)/2)
	startY := max(0, (height-n)/2)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			r, g, b, _ := img.At(bounds.Min.X+min(startX+x, width-1), bounds.Min.Y+min(startY+y, height-1)).RGBA()
			gray[y*n+x] = (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) * 255 / 65535
		}
	}
	re := make([]float64, n*n)
	im := make([]float64, n*n)
	copy(re, gray)
	for y := 0; y < n; y++ {
		fft1d(re[y*n:(y+1)*n], im[y*n:(y+1)*n])
	}
	for x := 0; x < n; x++ {
		rowRe := make([]float64, n)
		rowIm := make([]float64, n)
		for y := 0; y < n; y++ {
			rowRe[y] = re[y*n+x]
			rowIm[y] = im[y*n+x]
		}
		fft1d(rowRe, rowIm)
		for y := 0; y < n; y++ {
			re[y*n+x] = rowRe[y]
			im[y*n+x] = rowIm[y]
		}
	}
	magnitudes := make([]float64, n*n)
	for i := range magnitudes {
		magnitudes[i] = math.Log1p(cmplx.Abs(complex(re[i], im[i])))
	}
	shifted := make([]float64, n*n)
	half := n / 2
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			sx, sy := (x+half)%n, (y+half)%n
			shifted[sy*n+sx] = magnitudes[y*n+x]
		}
	}
	center := n / 2
	radius := float64(n) * 0.25
	total, high := 0.0, 0.0
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			energy := shifted[y*n+x] * shifted[y*n+x]
			total += energy
			dx, dy := float64(x-center), float64(y-center)
			if math.Hypot(dx, dy) > radius {
				high += energy
			}
		}
	}
	ratio := 0.0
	if total > 0 {
		ratio = high / total
	}
	anomalies := make([]string, 0)
	if ratio > 0.50 {
		anomalies = append(anomalies, "HIGH_FREQ_EXCESS")
	}
	if ratio < 0.08 {
		anomalies = append(anomalies, "LOW_FREQ_DEFICIT")
	}
	radial := make([]float64, center)
	counts := make([]float64, center)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			bin := min(int(math.Hypot(float64(x-center), float64(y-center))), center-1)
			radial[bin] += shifted[y*n+x]
			counts[bin]++
		}
	}
	for i := range radial {
		if counts[i] > 0 {
			radial[i] /= counts[i]
		}
	}
	peaks := 0
	for i := 3; i < center-2; i++ {
		local := (radial[i-2] + radial[i-1] + radial[i+1] + radial[i+2]) / 4
		if local > 0 && radial[i]/local > 3 {
			peaks++
		}
	}
	if peaks >= 3 {
		anomalies = append(anomalies, "PERIODIC_PEAKS")
	}
	ganScore := 1.0
	for _, anomaly := range anomalies {
		switch anomaly {
		case "PERIODIC_PEAKS":
			ganScore -= 0.40
		case "HIGH_FREQ_EXCESS":
			ganScore -= 0.25
		case "LOW_FREQ_DEFICIT":
			ganScore -= 0.20
		case "SPECTRAL_SYMMETRY":
			ganScore -= 0.15
		}
	}
	return frequencyResult{GanScore: math.Max(0, math.Min(1, ganScore)), EnergyRatio: ratio, Anomalies: anomalies}, nil
}

func fft1d(re, im []float64) {
	n := len(re)
	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for j&bit != 0 {
			j ^= bit
			bit >>= 1
		}
		j ^= bit
		if i < j {
			re[i], re[j] = re[j], re[i]
			im[i], im[j] = im[j], im[i]
		}
	}
	for length := 2; length <= n; length <<= 1 {
		half := length >> 1
		w := cmplx.Exp(complex(0, -2*math.Pi/float64(length)))
		for i := 0; i < n; i += length {
			cur := complex(1, 0)
			for j := 0; j < half; j++ {
				u := complex(re[i+j], im[i+j])
				v := complex(re[i+j+half], im[i+j+half]) * cur
				re[i+j], im[i+j] = real(u+v), imag(u+v)
				re[i+j+half], im[i+j+half] = real(u-v), imag(u-v)
				cur *= w
			}
		}
	}
}

func nextPow2(value int) int {
	p := 1
	for p < value {
		p <<= 1
	}
	return p
}
