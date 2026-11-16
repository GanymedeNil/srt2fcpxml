package convert

import (
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
)

func textSubs(cues ...struct {
	from, to time.Duration
	lines    []string
}) *astisub.Subtitles {
	s := astisub.NewSubtitles()
	for _, c := range cues {
		item := &astisub.Item{StartAt: c.from, EndAt: c.to}
		for _, l := range c.lines {
			item.Lines = append(item.Lines, astisub.Line{Items: []astisub.LineItem{{Text: l}}})
		}
		s.Items = append(s.Items, item)
	}
	return s
}

type tc = struct {
	from, to time.Duration
	lines    []string
}

func joinLines(lines []astisub.Line) string {
	var s []string
	for _, l := range lines {
		s = append(s, l.String())
	}
	return strings.Join(s, "|")
}

func TestPairSecondaryByLargestOverlap(t *testing.T) {
	sec := time.Second
	primary := textSubs(tc{0, 2 * sec, []string{"甲"}}, tc{2 * sec, 4 * sec, []string{"乙"}})
	secondary := textSubs(
		tc{100 * time.Millisecond, 1900 * time.Millisecond, []string{"A"}},
		tc{1500 * time.Millisecond, 3800 * time.Millisecond, []string{"B"}}, // overlaps 乙 more than 甲
		tc{3900 * time.Millisecond, 4500 * time.Millisecond, []string{"B2"}},
		tc{10 * sec, 11 * sec, []string{"orphan"}},
	)
	cues, err := buildCues(primary, Options{Secondary: secondary})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, c := range cues {
		got = append(got, joinLines(c.primary)+"/"+joinLines(c.secondary))
	}
	want := []string{"甲/A", "乙/B|B2", "/orphan"}
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Errorf("cues = %q, want %q", got, want)
	}
}

func TestSplitLines(t *testing.T) {
	subs := textSubs(tc{0, time.Second, []string{"中文", "English", "more"}}, tc{time.Second, 2 * time.Second, []string{"只有一行"}})
	cues, err := buildCues(subs, Options{SplitLines: true})
	if err != nil {
		t.Fatal(err)
	}
	if joinLines(cues[0].primary) != "中文" || joinLines(cues[0].secondary) != "English|more" {
		t.Errorf("first cue split wrong: %+v", cues[0])
	}
	if joinLines(cues[1].primary) != "只有一行" || len(cues[1].secondary) != 0 {
		t.Errorf("single-line cue should stay primary: %+v", cues[1])
	}
}

func TestBilingualOptionErrors(t *testing.T) {
	subs := textSubs(tc{0, time.Second, []string{"x"}})
	if _, err := buildCues(subs, Options{Secondary: subs, SplitLines: true}); err == nil {
		t.Error("secondary file and split lines together should be an error")
	}
	if _, err := buildCues(subs, Options{Secondary: astisub.NewSubtitles()}); err == nil {
		t.Error("empty secondary subtitles should be an error")
	}
}
