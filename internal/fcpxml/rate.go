package fcpxml

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
	// suffix is the frame rate part of Final Cut Pro's predefined format names.
	suffix string
}{
	{24000.0 / 1001, Rate{1001, 24000}, "2398"},
	{24, Rate{100, 2400}, "24"},
	{25, Rate{100, 2500}, "25"},
	{30000.0 / 1001, Rate{1001, 30000}, "2997"},
	{30, Rate{100, 3000}, "30"},
	{50, Rate{100, 5000}, "50"},
	{60000.0 / 1001, Rate{1001, 60000}, "5994"},
	{60, Rate{100, 6000}, "60"},
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

// Timebase returns the whole-number rate that timecode counts in (24 for 23.976).
func (r Rate) Timebase() int64 {
	return (r.FrameDen + r.FrameNum/2) / r.FrameNum
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

// Frame sizes that have a predefined Final Cut Pro format at every supported rate,
// mapped to the size part of the format name (from Apple's "Predefined Video Formats").
var predefinedSizes = map[[2]int]string{
	{1920, 1080}: "1080p",
	{1280, 720}:  "720p",
	{1280, 1080}: "1280x1080p",
	{1440, 1080}: "1440x1080p",
	{2048, 1024}: "2048x1024p",
	{2048, 1080}: "2048x1080p",
	{2048, 1152}: "2048x1152p",
	{2048, 1536}: "2048x1536p",
	{2048, 1556}: "2048x1556p",
	{3840, 2160}: "3840x2160p",
	{4096, 2048}: "4096x2048p",
	{4096, 2160}: "4096x2160p",
	{4096, 2304}: "4096x2304p",
	{4096, 3112}: "4096x3112p",
	{5120, 2160}: "5120x2160p",
	{5120, 2560}: "5120x2560p",
	{5120, 2700}: "5120x2700p",
	{960, 720}:   "960x720p",
}

// PredefinedFormatName returns Final Cut Pro's predefined name for a progressive
// format of this size and rate (e.g. "FFVideoFormat1080p2398"), or "" for a custom
// format such as a vertical video. The format element's width, height and
// frameDuration attributes take precedence over the name, so "" is always safe.
func PredefinedFormatName(r Rate, width, height int) string {
	size, ok := predefinedSizes[[2]int{width, height}]
	if !ok {
		return ""
	}
	for _, k := range supportedRates {
		if k.rate == r {
			return "FFVideoFormat" + size + k.suffix
		}
	}
	return ""
}

// RateFromFraction returns the supported rate of exactly num/den frames per
// second (as ffprobe reports it, e.g. 24000/1001), and false for any other rate.
func RateFromFraction(num, den int64) (Rate, bool) {
	for _, k := range supportedRates {
		// num/den == FrameDen/FrameNum
		if num > 0 && den > 0 && num*k.rate.FrameNum == den*k.rate.FrameDen {
			return k.rate, true
		}
	}
	return Rate{}, false
}

// IsNTSC reports whether the rate is one of the 1000/1001 rates (23.976, 29.97, 59.94).
func (r Rate) IsNTSC() bool {
	return r.FrameNum == 1001
}

// TimecodeFrames converts a timecode such as "01:00:00:00" to a frame count.
// A ';' before the frames ("00:59:59;28") marks drop-frame timecode, which is
// only valid at 29.97 and 59.94 fps.
func (r Rate) TimecodeFrames(tc string) (frames int64, dropFrame bool, err error) {
	var hh, mm, ss, ff int64
	sep := byte(':')
	if i := strings.LastIndexAny(tc, ":;."); i >= 0 {
		sep = tc[i]
	}
	if _, err := fmt.Sscanf(strings.NewReplacer(";", ":", ".", ":").Replace(tc), "%d:%d:%d:%d", &hh, &mm, &ss, &ff); err != nil {
		return 0, false, fmt.Errorf("invalid timecode %q", tc)
	}
	tb := r.Timebase()
	if mm > 59 || ss > 59 || ff >= tb || hh < 0 || mm < 0 || ss < 0 || ff < 0 {
		return 0, false, fmt.Errorf("invalid timecode %q at %v fps", tc, r.Timebase())
	}
	frames = (hh*3600+mm*60+ss)*tb + ff
	if sep == ':' {
		return frames, false, nil
	}
	if !r.IsNTSC() || tb%30 != 0 {
		return 0, false, fmt.Errorf("drop-frame timecode %q needs 29.97 or 59.94 fps", tc)
	}
	// Drop-frame skips frame numbers 0 and 1 (0-3 at 59.94) at the start of every
	// minute except each tenth minute.
	drop := tb / 15
	minutes := hh*60 + mm
	return frames - drop*(minutes-minutes/10), true, nil
}
