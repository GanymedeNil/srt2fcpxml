package convert

import (
	"bytes"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/input"
	"github.com/asticode/go-astisub"
)

var update = flag.Bool("update", false, "rewrite testdata/*.fcpxml golden files")

// fixedOptions returns options that produce byte-for-byte stable output.
func fixedOptions(t *testing.T, fps string) Options {
	t.Helper()
	rate, err := fcpxml.ParseRate(fps)
	if err != nil {
		t.Fatal(err)
	}
	uid := 0
	return Options{
		ProjectName:     "basic",
		Rate:            rate,
		Width:           1920,
		Height:          1080,
		Style:           DefaultStyle(),
		LibraryLocation: "file:///Users/test/Movies/basic.fcpbundle",
		Now:             time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("", 8*3600)),
		NewUID: func() string {
			uid++
			return fmt.Sprintf("00000000-0000-4000-8000-%012d", uid)
		},
	}
}

func TestGolden(t *testing.T) {
	custom, err := LoadStyle("testdata/custom.json", DefaultStyle())
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		srt       string
		secondary string
		split     bool
		fps       string
		width     int
		height    int
		style     Style
		video     *Video
	}{
		{name: "default-25", srt: "basic.srt", fps: "25", width: 1920, height: 1080, style: DefaultStyle()},
		{name: "default-23.976", srt: "basic.srt", fps: "23.976", width: 1920, height: 1080, style: DefaultStyle()},
		{name: "custom-vertical-29.97", srt: "basic.srt", fps: "29.97", width: 1080, height: 1920, style: custom},
		{name: "bilingual-files-25", srt: "basic.srt", secondary: "basic-en.srt", fps: "25", width: 1920, height: 1080, style: DefaultStyle()},
		{name: "bilingual-split-25", srt: "bilingual-split.srt", split: true, fps: "25", width: 1920, height: 1080, style: DefaultStyle()},
		{name: "video-23.976", srt: "basic.srt", fps: "23.976", width: 1920, height: 1080, style: DefaultStyle(), video: &Video{
			Name: "interview", URL: "file:///Users/test/Movies/interview.mov", Duration: time.Hour + 2*time.Second,
			Timecode: "01:00:00:00", HasAudio: true, AudioRate: 48000, AudioChannels: 2,
		}},
		{name: "video-29.97-df-vertical", srt: "basic.srt", fps: "29.97", width: 1080, height: 1920, style: DefaultStyle(), video: &Video{
			Name: "phone clip", URL: "file:///Users/test/Movies/phone%20clip.mp4", Duration: time.Hour + 2*time.Second,
			Timecode: "00:59:59;28",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subs, err := input.Open(filepath.Join("testdata", tt.srt))
			if err != nil {
				t.Fatal(err)
			}
			opts := fixedOptions(t, tt.fps)
			opts.Width, opts.Height, opts.Style, opts.SplitLines, opts.Video = tt.width, tt.height, tt.style, tt.split, tt.video
			if tt.secondary != "" {
				if opts.Secondary, err = input.Open(filepath.Join("testdata", tt.secondary)); err != nil {
					t.Fatal(err)
				}
			}
			doc, err := Convert(subs, opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := doc.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join("testdata", tt.name+".fcpxml")
			if *update {
				if err := os.WriteFile(path, got, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("%v (run go test ./internal/convert -update to create it)", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("output differs from %s (run go test ./internal/convert -update if the change is intended)\n%s", path, got)
			}
		})
	}
}

// The DTDs ship inside Final Cut Pro, so this only runs where it is installed.
const fcpDTDDir = "/Applications/Final Cut Pro.app/Contents/Frameworks/Interchange.framework/Versions/A/Resources"

func TestGoldenFilesMatchDTD(t *testing.T) {
	xmllint, err := exec.LookPath("xmllint")
	if err != nil {
		t.Skip("xmllint not installed")
	}
	if _, err := os.Stat(fcpDTDDir); err != nil {
		t.Skip("Final Cut Pro not installed")
	}
	golden, _ := filepath.Glob("testdata/*.fcpxml")
	if len(golden) == 0 {
		t.Fatal("no golden files")
	}
	tmp := t.TempDir() // xmllint cannot load a DTD from a path with spaces
	for _, version := range []string{"1.7", "1.14"} {
		dtd, err := os.ReadFile(filepath.Join(fcpDTDDir, "FCPXMLv"+strings.ReplaceAll(version, ".", "_")+".dtd"))
		if err != nil {
			t.Skipf("FCPXML %s DTD not found: %v", version, err)
		}
		dtdPath := filepath.Join(tmp, "fcpxml-"+version+".dtd")
		if err := os.WriteFile(dtdPath, dtd, 0o644); err != nil {
			t.Fatal(err)
		}
		for _, g := range golden {
			doc, err := os.ReadFile(g)
			if err != nil {
				t.Fatal(err)
			}
			// FCPXML 1.9 moved an asset's src into a media-rep element, so documents
			// with a video are valid FCPXML 1.7 but not 1.14.
			if version != fcpxml.Version && bytes.Contains(doc, []byte("<asset ")) {
				continue
			}
			// Each DTD fixes the version attribute; the other elements used are the same in 1.7 and 1.14.
			doc = bytes.Replace(doc, []byte(`version="`+fcpxml.Version+`"`), []byte(`version="`+version+`"`), 1)
			docPath := filepath.Join(tmp, filepath.Base(g))
			if err := os.WriteFile(docPath, doc, 0o644); err != nil {
				t.Fatal(err)
			}
			if out, err := exec.Command(xmllint, "--noout", "--dtdvalid", dtdPath, docPath).CombinedOutput(); err != nil {
				t.Errorf("%s is not valid FCPXML %s:\n%s", g, version, out)
			}
		}
	}
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

func convert(t *testing.T, fps string, subs *astisub.Subtitles) (*fcpxml.Sequence, *big.Rat) {
	t.Helper()
	doc, err := Convert(subs, fixedOptions(t, fps))
	if err != nil {
		t.Fatal(err)
	}
	return &doc.Library.Events[0].Projects[0].Sequence, seconds(t, doc.Resources.Formats[0].FrameDuration)
}

// A subtitle one hour in must still be placed within half a frame of its SRT time.
func TestNoDriftOverAnHour(t *testing.T) {
	for _, fps := range []string{"23.98", "29.97", "59.94", "25"} {
		seq, frame := convert(t, fps, subtitles([2]time.Duration{3599 * time.Second, 3601 * time.Second}))
		title := seq.Spine.Gaps[0].Titles[0]
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
		seq, frame := convert(t, fps, subtitles(
			[2]time.Duration{1234 * time.Millisecond, 5678 * time.Millisecond},
			[2]time.Duration{5678 * time.Millisecond, 3*time.Hour + 17*time.Millisecond},
		))
		gap := seq.Spine.Gaps[0]
		values := []string{seq.Duration, gap.Duration, gap.Start}
		for _, ti := range gap.Titles {
			values = append(values, ti.Offset, ti.Duration, ti.Start)
		}
		for _, v := range values {
			if q := new(big.Rat).Quo(seconds(t, v), frame); !q.IsInt() {
				t.Errorf("%s fps: %s is not a whole number of frames", fps, v)
			}
		}
	}
}

// The gap's local timeline starts at 01:00:00:00 at every rate.
func TestGapStartsAtOneHourTimecode(t *testing.T) {
	want := map[string]string{"25": "3600s", "23.976": "3603.6s", "29.97": "3603.6s", "60": "3600s"}
	for fps, w := range want {
		seq, _ := convert(t, fps, subtitles([2]time.Duration{0, time.Second}))
		got := seconds(t, seq.Spine.Gaps[0].Start)
		if got.Cmp(seconds(t, w)) != 0 {
			t.Errorf("%s fps: gap start = %ss, want %s", fps, got.FloatString(3), w)
		}
	}
}

// Subtitles that touch in the SRT must touch on the timeline: no overlap, no 1-frame hole.
func TestAdjacentSubtitlesShareABoundary(t *testing.T) {
	ms := time.Millisecond
	// At 25 fps these boundaries round so that start + rounded duration would overlap the next cue.
	seq, _ := convert(t, "25", subtitles(
		[2]time.Duration{20 * ms, 99 * ms},
		[2]time.Duration{99 * ms, 200 * ms},
		[2]time.Duration{200 * ms, 1001 * ms},
	))
	titles := seq.Spine.Gaps[0].Titles
	for i := 0; i+1 < len(titles); i++ {
		end := new(big.Rat).Add(seconds(t, titles[i].Offset), seconds(t, titles[i].Duration))
		next := seconds(t, titles[i+1].Offset)
		if end.Cmp(next) != 0 {
			t.Errorf("title %d ends at %ss but title %d starts at %ss", i+1, end.FloatString(3), i+2, next.FloatString(3))
		}
	}
}

func TestSequenceCoversLastSubtitle(t *testing.T) {
	seq, _ := convert(t, "25", subtitles(
		[2]time.Duration{0, 10 * time.Second},
		[2]time.Duration{2 * time.Second, 3 * time.Second}, // out of order: ends before the first one
	))
	if got := seq.Duration; got != "25000/2500s" {
		t.Errorf("sequence duration = %s, want 25000/2500s (10s)", got)
	}
}

func TestInvalidInput(t *testing.T) {
	one := subtitles([2]time.Duration{0, time.Second})
	tests := map[string]func(o *Options) *astisub.Subtitles{
		"no subtitles":   func(o *Options) *astisub.Subtitles { return astisub.NewSubtitles() },
		"nil subtitles":  func(o *Options) *astisub.Subtitles { return nil },
		"zero rate":      func(o *Options) *astisub.Subtitles { o.Rate = fcpxml.Rate{}; return one },
		"bad resolution": func(o *Options) *astisub.Subtitles { o.Width = 0; return one },
		"bad style":      func(o *Options) *astisub.Subtitles { o.Style.FontSize = 0; return one },
	}
	for name, setup := range tests {
		opts := fixedOptions(t, "25")
		subs := setup(&opts)
		if _, err := Convert(subs, opts); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestVideoTitlesUseClipTime(t *testing.T) {
	opts := fixedOptions(t, "25")
	opts.Video = &Video{Name: "v", URL: "file:///v.mov", Duration: 10 * time.Second, Timecode: "01:00:00:00"}
	var warnings []string
	opts.Warnf = func(format string, args ...any) { warnings = append(warnings, fmt.Sprintf(format, args...)) }
	doc, err := Convert(subtitles(
		[2]time.Duration{time.Second, 2 * time.Second},
		[2]time.Duration{9 * time.Second, 11 * time.Second}, // ends after the video
	), opts)
	if err != nil {
		t.Fatal(err)
	}
	seq := doc.Library.Events[0].Projects[0].Sequence
	clip := seq.Spine.AssetClips[0]
	if clip.Start != "9000000/2500s" || clip.Duration != "25000/2500s" || seq.Duration != "25000/2500s" || clip.TCFormat != "NDF" {
		t.Errorf("clip = %+v, sequence duration %s", clip, seq.Duration)
	}
	// 01:00:00:00 + 1s at 25 fps = frame 90025.
	if got := clip.Titles[0].Offset; got != "9002500/2500s" {
		t.Errorf("first title offset = %s, want 9002500/2500s", got)
	}
	if a := doc.Resources.Assets[0]; a.Src != "file:///v.mov" || a.Start != clip.Start || a.HasAudio != "" {
		t.Errorf("asset = %+v", a)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "1 subtitle(s) end after the video ends") {
		t.Errorf("warnings = %q", warnings)
	}
}

func TestVideoErrors(t *testing.T) {
	one := subtitles([2]time.Duration{0, time.Second})
	for name, v := range map[string]Video{
		"bad timecode":         {Name: "v", Duration: time.Second, Timecode: "soon"},
		"drop-frame at 25 fps": {Name: "v", Duration: time.Second, Timecode: "00:00:00;00"},
		"no frames":            {Name: "v", Duration: time.Millisecond},
	} {
		opts := fixedOptions(t, "25")
		opts.Video = &v
		if _, err := Convert(one, opts); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
