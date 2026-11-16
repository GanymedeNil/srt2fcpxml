package extract

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/convert"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/input"
)

func describe(cues []Cue) []string {
	var out []string
	for _, c := range cues {
		out = append(out, fmt.Sprintf("%v-%v %s %q", c.Start, c.End, c.Kind, c.Text))
	}
	return out
}

func check(t *testing.T, got []Cue, want ...string) {
	t.Helper()
	g := describe(got)
	if strings.Join(g, "\n") != strings.Join(want, "\n") {
		t.Errorf("cues:\n%s\nwant:\n%s", strings.Join(g, "\n"), strings.Join(want, "\n"))
	}
}

func TestStructures(t *testing.T) {
	res, err := ReadPath("testdata/structures.fcpxml", Options{})
	if err != nil {
		t.Fatal(err)
	}
	check(t, res.Cues,
		`2s-4s title "Anchored on clip"`,       // anchored to an asset-clip whose local time starts at 10s
		`5s-6s title "In secondary storyline"`, // spine shares the clip's timeline; the disabled title is skipped
		`10s-11s title "第一行\nSecond line"`,     // runs joined in order
		`12s-14s caption "A caption"`,          // captions too
		`14s-15s title "Audition pick"`,        // only the active pick
		`21s-23s title "Inside compound"`,      // compound clip media mapped through the ref-clip
		`24s-25s title "Cut by compound end"`,  // contained items are cut to the ref-clip
		`26s-27s title "On retimed clip"`,
	)
	if res.Project != "Main" {
		t.Errorf("project = %q", res.Project)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0], `clip "Slow" is retimed`) {
		t.Errorf("warnings = %q", res.Warnings)
	}
}

func TestOnly(t *testing.T) {
	res, err := ReadPath("testdata/structures.fcpxml", Options{Only: Caption})
	if err != nil {
		t.Fatal(err)
	}
	check(t, res.Cues, `12s-14s caption "A caption"`)
	if _, err := ReadPath("testdata/structures.fcpxml", Options{Only: "marker"}); err == nil {
		t.Error("unknown kind should be an error")
	}
}

func TestProjects(t *testing.T) {
	_, err := ReadPath("testdata/two-projects.fcpxml", Options{})
	if err == nil || !strings.Contains(err.Error(), `"First", "Second"`) {
		t.Errorf("expected an error listing the projects, got %v", err)
	}
	res, err := ReadPath("testdata/two-projects.fcpxml", Options{Project: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	check(t, res.Cues, `2s-3s title "two"`)
	if _, err := ReadPath("testdata/two-projects.fcpxml", Options{Project: "Third"}); err == nil {
		t.Error("unknown project should be an error")
	}
}

func TestBundle(t *testing.T) {
	res, err := ReadPath("testdata/bundle.fcpxmld", Options{})
	if err != nil {
		t.Fatal(err)
	}
	check(t, res.Cues, `1.001s-2.001s title "From a bundle"`) // text directly in <text>
}

func TestErrors(t *testing.T) {
	for name, doc := range map[string]string{
		"not fcpxml": `<xmeml/>`,
		"no project": `<fcpxml version="1.7"><resources/></fcpxml>`,
		"bad time":   `<fcpxml><project name="p"><sequence><spine><gap offset="soon" duration="1s"/></spine></sequence></project></fcpxml>`,
		"not xml":    `{}`,
	} {
		if _, err := Extract(strings.NewReader(doc), Options{}); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

// Subtitles converted by srt2fcpxml come back with the same text, within a frame.
func TestRoundTrip(t *testing.T) {
	type tc struct {
		name, srt, fps string
		split          bool
		video          *convert.Video
	}
	for _, c := range []tc{
		{name: "gap 23.976", srt: "basic.srt", fps: "23.976"},
		{name: "video with drop-frame timecode", srt: "basic.srt", fps: "29.97", video: &convert.Video{
			Name: "v", URL: "file:///v.mov", Duration: 2 * time.Hour, Timecode: "00:59:59;28"}},
		{name: "bilingual", srt: "bilingual-split.srt", fps: "25", split: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			subs, err := input.Open("../convert/testdata/" + c.srt)
			if err != nil {
				t.Fatal(err)
			}
			rate, _ := fcpxml.ParseRate(c.fps)
			doc, err := convert.Convert(subs, convert.Options{
				ProjectName: "p", Rate: rate, Width: 1920, Height: 1080,
				Style: convert.DefaultStyle(), SplitLines: c.split, Video: c.video,
			})
			if err != nil {
				t.Fatal(err)
			}
			data, err := doc.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			res, err := Extract(bytes.NewReader(data), Options{})
			if err != nil {
				t.Fatal(err)
			}
			if len(res.Cues) != len(subs.Items) {
				t.Fatalf("got %d cues, want %d", len(res.Cues), len(subs.Items))
			}
			frame := time.Duration(float64(time.Second) / rate.FPS())
			for i, item := range subs.Items {
				got := res.Cues[i]
				var lines []string
				for _, l := range item.Lines {
					lines = append(lines, l.String())
				}
				if want := strings.Join(lines, "\n"); got.Text != want {
					t.Errorf("cue %d text = %q, want %q", i+1, got.Text, want)
				}
				if (got.Start-item.StartAt).Abs() > frame || (got.End-item.EndAt).Abs() > frame {
					t.Errorf("cue %d at %v-%v, want %v-%v", i+1, got.Start, got.End, item.StartAt, item.EndAt)
				}
			}
		})
	}
}

func TestEncode(t *testing.T) {
	cues := []Cue{{Start: time.Second, End: 2500 * time.Millisecond, Text: "你好\nHello"}}
	srt, err := Encode(cues, ".srt")
	if err != nil {
		t.Fatal(err)
	}
	// astisub starts SRT with a UTF-8 byte order mark, which some players need to read non-ASCII text.
	if want := "\ufeff1\n00:00:01,000 --> 00:00:02,500\n你好\nHello\n"; !strings.HasPrefix(string(srt), want) {
		t.Errorf("srt = %q", srt)
	}
	vtt, err := Encode(cues, ".VTT")
	if err != nil || !strings.HasPrefix(string(vtt), "WEBVTT") || !strings.Contains(string(vtt), "00:00:01.000 --> 00:00:02.500") {
		t.Errorf("vtt = %q, err %v", vtt, err)
	}
	if _, err := Encode(cues, ".ass"); err == nil {
		t.Error("unsupported format should be an error")
	}
}
