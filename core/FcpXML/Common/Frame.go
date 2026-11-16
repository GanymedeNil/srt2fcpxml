package Common

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Rate is a video frame rate, stored as its exact frame duration of
// FrameNum/FrameDen seconds (e.g. 1001/24000s for 23.976 fps, 100/2500s for 25 fps).
//
// FCPXML expresses every time value as a rational number of seconds, and timing
// values should be a multiple of the timeline's frame duration (otherwise Final
// Cut Pro inserts gaps on import). All timing math is therefore done in whole
// frames and only formatted as rational seconds at the end.
type Rate struct {
	FrameNum int64
	FrameDen int64
}

// SupportedRates lists the frame rates accepted by ParseRate.
const SupportedRates = "23.976 (23.98), 24, 25, 29.97, 30, 50, 59.94, 60"

var supportedRates = []struct {
	fps  float64
	rate Rate
}{
	{24000.0 / 1001, Rate{1001, 24000}},
	{24, Rate{100, 2400}},
	{25, Rate{100, 2500}},
	{30000.0 / 1001, Rate{1001, 30000}},
	{30, Rate{100, 3000}},
	{50, Rate{100, 5000}},
	{60000.0 / 1001, Rate{1001, 60000}},
	{60, Rate{100, 6000}},
}

// ParseRate parses a frame rate such as "25", "25.0", "23.98" or "23.976".
// NTSC rates may be written with two or three decimals; anything that is not
// one of SupportedRates is rejected.
func ParseRate(s string) (Rate, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err == nil {
		for _, k := range supportedRates {
			if math.Abs(v-k.fps) < 0.006 {
				return k.rate, nil
			}
		}
	}
	return Rate{}, fmt.Errorf("unsupported frame rate %q (supported: %s)", s, SupportedRates)
}

// IsZero reports whether the rate is unset.
func (r Rate) IsZero() bool {
	return r.FrameNum <= 0 || r.FrameDen <= 0
}

// FPS returns the number of frames per second.
func (r Rate) FPS() float64 {
	return float64(r.FrameDen) / float64(r.FrameNum)
}

// Frames converts a duration to the nearest whole number of frames.
func (r Rate) Frames(d time.Duration) int64 {
	// frames = d / (FrameNum/FrameDen s) = ns * FrameDen / (FrameNum * 1e9), rounded half up.
	n := new(big.Int).Mul(big.NewInt(d.Nanoseconds()), big.NewInt(r.FrameDen))
	q := new(big.Int).Mul(big.NewInt(r.FrameNum), big.NewInt(int64(time.Second)))
	n.Add(n, new(big.Int).Rsh(q, 1))
	return n.Quo(n, q).Int64()
}

// Time formats a whole number of frames as an FCPXML time value, e.g. "86486400/24000s".
func (r Rate) Time(frames int64) string {
	return fmt.Sprintf("%d/%ds", frames*r.FrameNum, r.FrameDen)
}

// FrameDuration returns the value of the format element's frameDuration attribute.
func (r Rate) FrameDuration() string {
	return r.Time(1)
}

// GapStartFrames is the local start of the gap that the titles are anchored to.
// Titles use it as their time base, so it must be the same for the gap and every
// title. The value (01:00:00:00 NDF for NTSC rates, but 00:06:00:00 for whole
// rates) is kept from the original 10.4.6 template so the timeline is unchanged.
func (r Rate) GapStartFrames() int64 {
	return r.FrameDen * 36 / 10
}
