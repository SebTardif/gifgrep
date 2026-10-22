package gifdecode

// maxCumulativePixels is the total pixel budget across every frame that
// will be decoded. It is large enough for more than the default 60 frames
// and still rejects a multi-gigabyte expansion of a small file.
const maxCumulativePixels int64 = 1 << 32

func limitGIFFrames(data []byte, maxFrames int) ([]byte, int, int, int, bool) {
	frames, width, height, ends, ok := scanGIFFrames(data)
	if !ok || frames == 0 {
		return data, 0, 0, 0, false
	}
	keep := frames
	if maxFrames > 0 && maxFrames < keep {
		keep = maxFrames
		trimmed := append([]byte{}, data[:ends[keep-1]]...)
		trimmed = append(trimmed, 0x3b)
		return trimmed, keep, width, height, true
	}
	return data, frames, width, height, true
}

func cumulativeTooLarge(frames, width, height int) bool {
	if frames <= 0 || width <= 0 || height <= 0 {
		return false
	}
	return int64(width)*int64(height)*int64(frames) > maxCumulativePixels
}

func scanGIFFrames(data []byte) (frames, width, height int, ends []int, ok bool) {
	if len(data) < 13 || (string(data[:6]) != "GIF87a" && string(data[:6]) != "GIF89a") {
		return 0, 0, 0, nil, false
	}
	width = int(data[6]) | int(data[7])<<8
	height = int(data[8]) | int(data[9])<<8
	i := 13
	if data[10]&0x80 != 0 {
		i += 3 * (1 << (int(data[10]&7) + 1))
	}
	for i < len(data) {
		switch data[i] {
		case 0x3b:
			return frames, width, height, ends, true
		case 0x21:
			i += 2
			i = skipGIFSubBlocks(data, i)
		case 0x2c:
			if i+10 > len(data) {
				return 0, 0, 0, nil, false
			}
			packed := data[i+9]
			i += 10
			if packed&0x80 != 0 {
				i += 3 * (1 << (int(packed&7) + 1))
			}
			if i >= len(data) {
				return 0, 0, 0, nil, false
			}
			i++
			next := skipGIFSubBlocks(data, i)
			if next < i {
				return 0, 0, 0, nil, false
			}
			i = next
			frames++
			ends = append(ends, i)
		default:
			return 0, 0, 0, nil, false
		}
	}
	return frames, width, height, ends, frames > 0
}

func skipGIFSubBlocks(data []byte, i int) int {
	for i < len(data) {
		size := int(data[i])
		i++
		if size == 0 {
			return i
		}
		if i+size > len(data) {
			return len(data) + 1
		}
		i += size
	}
	return len(data) + 1
}
