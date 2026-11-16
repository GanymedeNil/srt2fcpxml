package core

import (
	"encoding/xml"
	"math/big"
	"srt2fcpxml/core/FcpXML/Common"
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
)

type parsedTitle struct {
	Offset   string `xml:"offset,attr"`
	Duration string `xml:"duration,attr"`
	Start    string `xml:"start,attr"`
}

type parsedDoc struct {
	Format struct {
		FrameDuration string `xml:"frameDuration,attr"`
	} `xml:"resources>format"`
	Sequence struct {
		Duration string `xml:"duration,attr"`
		Gap      struct {
			Duration string        `xml:"duration,attr"`
			Start    string        `xml:"start,attr"`
			Titles   []parsedTitle `xml:"title"`
		} `xml:"spine>gap"`
	} `xml:"library>event>project>sequence"`
}

// seconds parses an FCPXML time value such as "86486400/24000s" or "0s".
func seconds(t *testing.T, v string) *big.Rat {
	t.Helper()
	r, ok := new(big.Rat).SetString(strings.TrimSuffix(v, "s"))
	if !ok {
		t.Fatalf("bad time value %q", v)
	}
	return r
}

func subtitles(cues ...[2]time.Duration) *astisub.Subtitles {
	s := astisub.NewSubtitles()
	for _, c := range cues {
		s.Items = append(s.Items, &astisub.Item{
			StartAt: c[0],
			EndAt:   c[1],
			Lines:   []astisub.Line{{Items: []astisub.LineItem{{Text: "x"}}}},
		})
	}
	return s
}

func export(t *testing.T, fps string, subs *astisub.Subtitles) parsedDoc {
	t.Helper()
	rate, err := Common.ParseRate(fps)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Srt2FcpXmlExport("test", rate, subs, 1920, 1080)
	if err != nil {
		t.Fatal(err)
	}
	var doc parsedDoc
	if err := xml.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// A subtitle one hour in must still be placed within half a frame of its SRT time.
func TestNoDriftOverAnHour(t *testing.T) {
	for _, fps := range []string{"23.98", "29.97", "59.94", "25"} {
		doc := export(t, fps, subtitles([2]time.Duration{3599 * time.Second, 3601 * time.Second}))
		title := doc.Sequence.Gap.Titles[0]
		frame := seconds(t, doc.Format.FrameDuration)
		got := new(big.Rat).Sub(seconds(t, title.Offset), seconds(t, title.Start))
		diff := new(big.Rat).Sub(got, big.NewRat(3599, 1))
		halfFrame := new(big.Rat).Mul(frame, big.NewRat(1, 2))
		if diff.Abs(diff).Cmp(halfFrame) > 0 {
			t.Errorf("%s fps: subtitle at 3599s placed at %ss", fps, got.FloatString(4))
		}
	}
}

// Every time value must be a whole number of frames, or Final Cut Pro inserts gaps on import.
func TestTimesAreFrameAligned(t *testing.T) {
	for _, fps := range []string{"23.98", "24", "25", "29.97", "30", "50", "59.94", "60"} {
		doc := export(t, fps, subtitles(
			[2]time.Duration{1234 * time.Millisecond, 5678 * time.Millisecond},
			[2]time.Duration{5678 * time.Millisecond, 3*time.Hour + 17*time.Millisecond},
		))
		frame := seconds(t, doc.Format.FrameDuration)
		values := []string{doc.Sequence.Duration, doc.Sequence.Gap.Duration, doc.Sequence.Gap.Start}
		for _, ti := range doc.Sequence.Gap.Titles {
			values = append(values, ti.Offset, ti.Duration, ti.Start)
		}
		for _, v := range values {
			if q := new(big.Rat).Quo(seconds(t, v), frame); !q.IsInt() {
				t.Errorf("%s fps: %s is not a whole number of frames", fps, v)
			}
		}
	}
}

// Subtitles that touch in the SRT must touch on the timeline: no overlap, no 1-frame hole.
func TestAdjacentSubtitlesShareABoundary(t *testing.T) {
	ms := time.Millisecond
	// At 25 fps these boundaries round so that start + rounded duration would overlap the next cue.
	doc := export(t, "25", subtitles(
		[2]time.Duration{20 * ms, 99 * ms},
		[2]time.Duration{99 * ms, 200 * ms},
		[2]time.Duration{200 * ms, 1001 * ms},
	))
	titles := doc.Sequence.Gap.Titles
	for i := 0; i+1 < len(titles); i++ {
		end := new(big.Rat).Add(seconds(t, titles[i].Offset), seconds(t, titles[i].Duration))
		next := seconds(t, titles[i+1].Offset)
		if end.Cmp(next) != 0 {
			t.Errorf("title %d ends at %ss but title %d starts at %ss", i+1, end.FloatString(3), i+2, next.FloatString(3))
		}
	}
}

func TestSequenceCoversLastSubtitle(t *testing.T) {
	doc := export(t, "25", subtitles(
		[2]time.Duration{0, 10 * time.Second},
		[2]time.Duration{2 * time.Second, 3 * time.Second}, // out of order: ends before the first one
	))
	if got := doc.Sequence.Duration; got != "25000/2500s" {
		t.Errorf("sequence duration = %s, want 25000/2500s (10s)", got)
	}
}

func TestInvalidInput(t *testing.T) {
	rate, _ := Common.ParseRate("25")
	one := subtitles([2]time.Duration{0, time.Second})
	tests := map[string]func() error{
		"no subtitles":   func() error { _, err := Srt2FcpXmlExport("p", rate, astisub.NewSubtitles(), 1920, 1080); return err },
		"nil subtitles":  func() error { _, err := Srt2FcpXmlExport("p", rate, nil, 1920, 1080); return err },
		"zero rate":      func() error { _, err := Srt2FcpXmlExport("p", Common.Rate{}, one, 1920, 1080); return err },
		"bad resolution": func() error { _, err := Srt2FcpXmlExport("p", rate, one, 0, 1080); return err },
	}
	for name, f := range tests {
		if f() == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
