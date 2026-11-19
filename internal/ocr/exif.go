package ocr

import "encoding/binary"

// exifOrientation 解析 JPEG EXIF Orientation（tag 0x0112），返回需顺时针旋转的角度；
// 无 EXIF、非 JPEG 或含镜像变换（2/4/5/7）时返回 0，交由原方向处理。
func exifOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0
	}
	pos := 2
	for pos+4 <= len(data) {
		if data[pos] != 0xFF {
			return 0
		}
		marker := data[pos+1]
		if marker == 0xD8 || marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			pos += 2
			continue
		}
		if marker == 0xDA || marker == 0xD9 {
			return 0
		}
		if pos+4 > len(data) {
			return 0
		}
		segLen := int(binary.BigEndian.Uint16(data[pos+2 : pos+4]))
		if segLen < 2 || pos+2+segLen > len(data) {
			return 0
		}
		if marker == 0xE1 && segLen >= 10 && string(data[pos+4:pos+10]) == "Exif\x00\x00" {
			return tiffOrientation(data[pos+10 : pos+2+segLen])
		}
		pos += 2 + segLen
	}
	return 0
}

// tiffOrientation 从 TIFF/Exif IFD0 读取 Orientation 短整型值。
func tiffOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 0
	}
	var bo binary.ByteOrder = binary.LittleEndian
	switch string(tiff[0:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 0
	}
	off := int(bo.Uint32(tiff[4:8]))
	if off <= 0 || off+2 > len(tiff) {
		return 0
	}
	n := int(bo.Uint16(tiff[off : off+2]))
	for i := 0; i < n; i++ {
		entry := tiff[off+2+i*12:]
		if len(entry) < 12 {
			return 0
		}
		if bo.Uint16(entry[0:2]) == 0x0112 && bo.Uint16(entry[2:4]) == 3 && bo.Uint32(entry[4:8]) == 1 {
			o := int(bo.Uint16(entry[8:10]))
			if o >= 1 && o <= 8 {
				return o
			}
			return 0
		}
	}
	return 0
}

// exifRotationAngle 把 EXIF Orientation 映射为顺时针旋转角度；镜像类方向不支持，返回 0。
func exifRotationAngle(orientation int) int {
	switch orientation {
	case 3:
		return 180
	case 6:
		return 90
	case 8:
		return 270
	default:
		return 0
	}
}
