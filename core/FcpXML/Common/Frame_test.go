package Common

import (
	"math"
	"testing"
	"time"
)

func TestParseRate(t *testing.T) {
	tests := []struct {
		in      string
		want    Rate
		wantErr bool
	}{
		{in: "23.98", want: Rate{1001, 24000}},
		{in: "23.976", want: Rate{1001, 24000}},
		{in: "24", want: Rate{100, 2400}},
		{in: "25", want: Rate{100, 2500}},
		{in: "25.0", want: Rate{100, 2500}},
		{in: " 29.97 ", want: Rate{1001, 30000}},
		{in: "30", want: Rate{100, 3000}},
		{in: "50", want: Rate{100, 5000}},
		{in: "59.94", want: Rate{1001, 60000}},
		{in: "60", want: Rate{100, 6000}},
		{in: "120", wantErr: true},
		{in: "23.9", wantErr: true},
		{in: "0", wantErr: true},
		{in: "-25", wantErr: true},
		{in: "abc", wantErr: true},
		{in: "NaN", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseRate(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRate(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseRate(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFrameDuration(t *testing.T) {
	tests := map[string]string{
		"23.98": "1001/24000s",
		"25":    "100/2500s",
		"29.97": "1001/30000s",
		"60":    "100/6000s",
	}
	for in, want := range tests {
		r, err := ParseRate(in)
		if err != nil {
			t.Fatal(err)
		}
		if got := r.FrameDuration(); got != want {
			t.Errorf("%s: FrameDuration() = %s, want %s", in, got, want)
		}
	}
}

// Every time value must map to the nearest frame, however far into the video it is.
// Converting with a rounded rate (23.98 instead of 24000/1001) drifts ~0.6s per hour.
func TestFramesDoNotDrift(t *testing.T) {
	for _, k := range supportedRates {
		r := k.rate
		frameSeconds := float64(r.FrameNum) / float64(r.FrameDen)
		for d := time.Duration(0); d <= 10*time.Hour; d += 7*time.Minute + 1234567*time.Microsecond {
			frames := r.Frames(d)
			got := float64(frames) * frameSeconds
			if diff := math.Abs(got - d.Seconds()); diff > frameSeconds/2+1e-9 {
				t.Fatalf("%v fps: %v -> frame %d (%.4fs), off by %.4fs", k.fps, d, frames, got, diff)
			}
		}
	}
}

func TestFramesKnownValues(t *testing.T) {
	ntsc24, _ := ParseRate("23.98")
	pal, _ := ParseRate("25")
	tests := []struct {
		rate Rate
		d    time.Duration
		want int64
	}{
		{pal, 0, 0},
		{pal, 1 * time.Second, 25},
		{pal, 19 * time.Millisecond, 0},     // 0.475 frames
		{pal, 20 * time.Millisecond, 1},     // exactly half a frame rounds up
		{ntsc24, 1 * time.Second, 24},       // 23.976 frames
		{ntsc24, 3599 * time.Second, 86290}, // 86289.71 frames
	}
	for _, tt := range tests {
		if got := tt.rate.Frames(tt.d); got != tt.want {
			t.Errorf("%v.Frames(%v) = %d, want %d", tt.rate, tt.d, got, tt.want)
		}
	}
}

func TestGapStartFrames(t *testing.T) {
	ntsc24, _ := ParseRate("23.976")
	pal, _ := ParseRate("25")
	if got := ntsc24.Time(ntsc24.GapStartFrames()); got != "86486400/24000s" {
		t.Errorf("23.976 gap start = %s", got)
	}
	if got := pal.Time(pal.GapStartFrames()); got != "900000/2500s" {
		t.Errorf("25 gap start = %s", got)
	}
}
