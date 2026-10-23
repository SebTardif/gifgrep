package gifdecode

import "testing"

func TestMaxFramesIgnoresMalformedTail(t *testing.T) {
	data := makeTestGIF(3)
	data[len(data)-1] = 0
	opts := DefaultOptions()
	opts.MaxFrames = 2
	opts.StrictGIF = true
	frames, err := Decode(data, opts)
	if err != nil {
		t.Fatalf("decode capped animation: %v", err)
	}
	if len(frames.Frames) != 2 {
		t.Fatalf("got %d frames, want 2", len(frames.Frames))
	}
}

func TestNegativeMaxFramesReturnsMoreThanDefault(t *testing.T) {
	opts := DefaultOptions()
	opts.MaxFrames = -1
	frames, err := Decode(makeTestGIF(70), opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames.Frames) != 70 {
		t.Fatalf("got %d frames, want 70", len(frames.Frames))
	}
}

func TestFrameCapRejectsMalformedRetainedData(t *testing.T) {
	valid := makeTestGIF(3)
	for end := 6; end < len(valid)-1; end++ {
		opts := DefaultOptions()
		opts.MaxFrames = 4
		opts.StrictGIF = true
		if _, err := Decode(valid[:end], opts); err == nil {
			t.Fatalf("truncated at %d: expected an error", end)
		}
	}
}

func TestFrameCapMatchesCompleteDecode(t *testing.T) {
	for _, frameLimit := range []int{1, 2, 3, 4} {
		opts := DefaultOptions()
		opts.MaxFrames = frameLimit
		got, err := Decode(makeTestGIF(3), opts)
		if err != nil {
			t.Fatal(err)
		}
		want := min(frameLimit, 3)
		if len(got.Frames) != want {
			t.Fatalf("cap %d: got %d frames, want %d", frameLimit, len(got.Frames), want)
		}
	}
}
