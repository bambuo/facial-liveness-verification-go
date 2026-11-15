package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

// ReadWAV 读取 RIFF/WAVE 文件，支持 PCM16 与 IEEE float32，返回单声道样本。
func ReadWAV(path string) (int, []float32, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}
	fileSize := stat.Size()
	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return 0, nil, err
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return 0, nil, fmt.Errorf("wav: invalid RIFF/WAVE header")
	}
	var (
		sampleRate int
		channels   int
		bits       int
		format     int
		data       []byte
	)
	for {
		chunkHeader := make([]byte, 8)
		if _, err := io.ReadFull(f, chunkHeader); err != nil {
			if err == io.EOF {
				break
			}
			return 0, nil, err
		}
		id := string(chunkHeader[0:4])
		size64 := int64(binary.LittleEndian.Uint32(chunkHeader[4:8]))
		if size64 < 0 || size64 > fileSize {
			return 0, nil, fmt.Errorf("wav: invalid chunk size %d", size64)
		}
		size := int(size64)
		chunk := make([]byte, size)
		if _, err := io.ReadFull(f, chunk); err != nil {
			return 0, nil, err
		}
		if size%2 != 0 {
			_, _ = f.Seek(1, io.SeekCurrent)
		}
		switch id {
		case "fmt ":
			if len(chunk) < 16 {
				return 0, nil, fmt.Errorf("wav: short fmt chunk")
			}
			format = int(binary.LittleEndian.Uint16(chunk[0:2]))
			channels = int(binary.LittleEndian.Uint16(chunk[2:4]))
			sampleRate = int(binary.LittleEndian.Uint32(chunk[4:8]))
			bits = int(binary.LittleEndian.Uint16(chunk[14:16]))
		case "data":
			data = chunk
		}
	}
	if sampleRate <= 0 || channels <= 0 || len(data) == 0 {
		return 0, nil, fmt.Errorf("wav: missing fmt/data chunk")
	}
	var mono []float32
	switch {
	case format == 1 && bits == 16:
		n := len(data) / 2
		frames := n / channels
		mono = make([]float32, frames)
		for i := 0; i < frames; i++ {
			var sum float64
			for c := 0; c < channels; c++ {
				off := (i*channels + c) * 2
				sum += float64(int16(binary.LittleEndian.Uint16(data[off:off+2]))) / 32768
			}
			mono[i] = float32(sum / float64(channels))
		}
	case format == 3 && bits == 32:
		n := len(data) / 4
		frames := n / channels
		mono = make([]float32, frames)
		for i := 0; i < frames; i++ {
			var sum float64
			for c := 0; c < channels; c++ {
				off := (i*channels + c) * 4
				sum += float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off : off+4])))
			}
			mono[i] = float32(sum / float64(channels))
		}
	default:
		return 0, nil, fmt.Errorf("wav: unsupported format=%d bits=%d", format, bits)
	}
	return sampleRate, mono, nil
}
