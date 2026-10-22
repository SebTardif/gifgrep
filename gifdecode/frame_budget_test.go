package gifdecode

import (
	"bytes"
	"errors"
	"testing"
)

func TestMaxFramesIgnoresFramesPastTheCap(t *testing.T) {
	data := makeTestGIF(3)
	scanned, width, height, ends, ok := scanGIFFrames(data)
	if !ok || len(ends) < 2 || scanned < 2 || width <= 0 || height <= 0 {
		t.Fatal("could not find the second frame")
	}
	broken := append([]byte{}, data[:ends[1]]...)
	broken = append(broken, 0x2c, 0, 0, 0, 0, 1, 0, 1, 0, 0, 2, 0)

	opts := DefaultOptions()
	opts.MaxFrames = 2
	opts.StrictGIF = true
	frames, err := Decode(broken, opts)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) != 2 {
		t.Fatalf("got %d frames", len(frames.Frames))
	}
}

func TestNegativeMaxFramesReturnsMoreThanDefault(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxFrames = -1
	frames, err := Decode(makeTestGIF(70), opts)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) != 70 {
		t.Fatalf("got %d frames, want 70", len(frames.Frames))
	}
}

func TestCumulativePixelsRejectsBeforeDecode(t *testing.T) {
	data := syntheticGIF(40, 20000, 20000)
	opts := DefaultOptions()
	opts.MaxFrames = -1
	opts.MaxPixels = 500_000_000
	_, err := Decode(data, opts)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("got %v, want ErrTooLarge", err)
	}
}

func syntheticGIF(frames, width, height int) []byte {
	var buf bytes.Buffer
	buf.WriteString("GIF89a")
	buf.Write([]byte{byte(width), byte(width >> 8), byte(height), byte(height >> 8), 0x80, 0, 0})
	buf.Write([]byte{0, 0, 0, 255, 255, 255})
	for i := 0; i < frames; i++ {
		buf.WriteByte(0x2c)
		buf.Write([]byte{0, 0, 0, 0, byte(width), byte(width >> 8), byte(height), byte(height >> 8), 0})
		buf.WriteByte(2)
		buf.WriteByte(1)
		buf.WriteByte(0)
		buf.WriteByte(0)
	}
	buf.WriteByte(0x3b)
	return buf.Bytes()
}
