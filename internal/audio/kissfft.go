package audio

import "math"

// kissCpx 是 KissFFT 的 float32 复数类型。
type kissCpx struct {
	r, i float32
}

// rfftPower 计算 KissFFT 兼容的实数 FFT power spectrum，返回 paddedSize/2 个 bin。
func rfftPower(frame []float32, paddedSize int) []float32 {
	ncfft := paddedSize / 2
	in := make([]kissCpx, ncfft)
	for i := 0; i < ncfft; i++ {
		in[i] = kissCpx{frame[2*i], frame[2*i+1]}
	}
	twiddles := kissTwiddles(ncfft)
	factors := kissFactors(ncfft)
	tmp := make([]kissCpx, ncfft)
	kfWork(tmp, in, 1, 1, factors, ncfft, twiddles)

	out := make([]kissCpx, ncfft+1)
	tdc := tmp[0]
	out[0] = kissCpx{tdc.r + tdc.i, 0}
	out[ncfft] = kissCpx{tdc.r - tdc.i, 0}
	super := kissSuperTwiddles(ncfft)
	for k := 1; k <= ncfft/2; k++ {
		fpk := tmp[k]
		fpnk := kissCpx{tmp[ncfft-k].r, -tmp[ncfft-k].i}
		f1k := kissCpx{fpk.r + fpnk.r, fpk.i + fpnk.i}
		f2k := kissCpx{fpk.r - fpnk.r, fpk.i - fpnk.i}
		tw := kissMul(f2k, super[k-1])
		out[k] = kissCpx{0.5 * (f1k.r + tw.r), 0.5 * (f1k.i + tw.i)}
		out[ncfft-k] = kissCpx{0.5 * (f1k.r - tw.r), 0.5 * (tw.i - f1k.i)}
	}
	power := make([]float32, ncfft)
	for i := 0; i < ncfft; i++ {
		power[i] = out[i].r*out[i].r + out[i].i*out[i].i
	}
	return power
}

func kissTwiddles(n int) []kissCpx {
	tw := make([]kissCpx, n)
	for i := 0; i < n; i++ {
		phase := -2 * math.Pi * float64(i) / float64(n)
		tw[i] = kissCpx{float32(math.Cos(phase)), float32(math.Sin(phase))}
	}
	return tw
}

func kissSuperTwiddles(ncfft int) []kissCpx {
	tw := make([]kissCpx, ncfft/2)
	for i := 0; i < ncfft/2; i++ {
		phase := -math.Pi * (float64(i+1)/float64(ncfft) + 0.5)
		tw[i] = kissCpx{float32(math.Cos(phase)), float32(math.Sin(phase))}
	}
	return tw
}

func kissFactors(n int) []int {
	var factors []int
	p := 4
	floorSqrt := int(math.Floor(math.Sqrt(float64(n))))
	for {
		for n%p != 0 {
			switch p {
			case 4:
				p = 2
			case 2:
				p = 3
			default:
				p += 2
			}
			if p > floorSqrt {
				p = n
			}
		}
		n /= p
		factors = append(factors, p, n)
		if n <= 1 {
			break
		}
	}
	return factors
}

func kfWork(fout, f []kissCpx, fstride, inStride int, factors []int, nfft int, tw []kissCpx) {
	p, m := factors[0], factors[1]
	if m == 1 {
		for k := 0; k < p; k++ {
			fout[k] = f[k*fstride*inStride]
		}
	} else {
		for k := 0; k < p; k++ {
			kfWork(fout[k*m:(k+1)*m], f[k*fstride*inStride:], fstride*p, inStride, factors[2:], nfft, tw)
		}
	}
	switch p {
	case 2:
		kfBfly2(fout, fstride, m, tw)
	case 4:
		kfBfly4(fout, fstride, m, tw)
	default:
		panic("kissfft: unsupported radix")
	}
}

func kfBfly2(fout []kissCpx, fstride, m int, tw []kissCpx) {
	tw1 := 0
	for k := 0; k < m; k++ {
		t := kissMul(fout[m+k], tw[tw1])
		tw1 += fstride
		x := fout[k]
		fout[m+k] = kissCpx{x.r - t.r, x.i - t.i}
		fout[k] = kissCpx{x.r + t.r, x.i + t.i}
	}
}

func kfBfly4(fout []kissCpx, fstride, m int, tw []kissCpx) {
	m2, m3 := 2*m, 3*m
	tw1, tw2, tw3 := 0, 0, 0
	for k := 0; k < m; k++ {
		s0 := kissMul(fout[m+k], tw[tw1])
		s1 := kissMul(fout[m2+k], tw[tw2])
		s2 := kissMul(fout[m3+k], tw[tw3])
		tw1 += fstride
		tw2 += fstride * 2
		tw3 += fstride * 3
		base := fout[k]
		s5 := kissCpx{base.r - s1.r, base.i - s1.i}
		fout[k] = kissCpx{base.r + s1.r, base.i + s1.i}
		s3 := kissCpx{s0.r + s2.r, s0.i + s2.i}
		s4 := kissCpx{s0.r - s2.r, s0.i - s2.i}
		fout[m2+k] = kissCpx{fout[k].r - s3.r, fout[k].i - s3.i}
		fout[k] = kissCpx{fout[k].r + s3.r, fout[k].i + s3.i}
		fout[m+k] = kissCpx{s5.r + s4.i, s5.i - s4.r}
		fout[m3+k] = kissCpx{s5.r - s4.i, s5.i + s4.r}
	}
}

func kissMul(a, b kissCpx) kissCpx {
	return kissCpx{
		r: a.r*b.r - a.i*b.i,
		i: a.r*b.i + a.i*b.r,
	}
}
