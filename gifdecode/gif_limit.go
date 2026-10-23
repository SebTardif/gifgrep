package gifdecode

import (
	"bytes"
	"fmt"
	"io"
)

// limitGIFFrames stops at the requested frame before the image decoder allocates
// any pixels. Negative caps preserve complete-animation decoding.
func limitGIFFrames(data []byte, maxFrames int) ([]byte, error) {
	if maxFrames <= 0 || len(data) < 6 || (!bytes.Equal(data[:6], []byte("GIF87a")) && !bytes.Equal(data[:6], []byte("GIF89a"))) {
		return data, nil
	}
	if len(data) < 13 {
		return nil, io.ErrUnexpectedEOF
	}
	i := 13
	if data[10]&0x80 != 0 {
		i += 3 * (1 << (int(data[10]&7) + 1))
	}
	frames := 0
	for i < len(data) {
		switch data[i] {
		case 0x3b:
			return data, nil
		case 0x21:
			if i+2 > len(data) {
				return nil, io.ErrUnexpectedEOF
			}
			i += 2
		case 0x2c:
			if i+10 > len(data) {
				return nil, io.ErrUnexpectedEOF
			}
			packed := data[i+9]
			i += 10
			if packed&0x80 != 0 {
				i += 3 * (1 << (int(packed&7) + 1))
			}
			i++ // LZW minimum code size; the decoder validates its value.
			frames++
		default:
			return nil, fmt.Errorf("gif: unknown block type: 0x%.2x", data[i])
		}
		var err error
		i, err = skipGIFSubBlocks(data, i)
		if err != nil {
			return nil, err
		}
		if frames == maxFrames {
			trimmed := make([]byte, i+1)
			copy(trimmed, data[:i])
			trimmed[i] = 0x3b
			return trimmed, nil
		}
	}
	return nil, io.ErrUnexpectedEOF
}

func skipGIFSubBlocks(data []byte, i int) (int, error) {
	for i < len(data) {
		size := int(data[i])
		i++
		if size == 0 {
			return i, nil
		}
		if size > len(data)-i {
			return 0, io.ErrUnexpectedEOF
		}
		i += size
	}
	return 0, io.ErrUnexpectedEOF
}
