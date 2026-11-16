package input

import (
	"strings"
	"testing"
)

func lineTexts(t *testing.T, name, content string) []string {
	t.Helper()
	subs, err := Read(name, strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, item := range subs.Items {
		for _, l := range item.Lines {
			out = append(out, l.String())
		}
	}
	return out
}

func TestSRTKeepsSpacesBetweenTags(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:02,000\n<i>a</i> <b>b</b>  <font color=\"#FF0000\">c</font>\n<i>x</i>y\n"
	got := lineTexts(t, "a.srt", srt)
	want := []string{"a b  c", "xy"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("lines = %q, want %q", got, want)
	}
}

func TestSRTKeepsInlineStyles(t *testing.T) {
	srt := "1\n00:00:01,000 --> 00:00:02,000\n<i>a</i> <b>b</b>\n"
	subs, err := Read("a.SRT", strings.NewReader(srt))
	if err != nil {
		t.Fatal(err)
	}
	items := subs.Items[0].Lines[0].Items
	if len(items) != 3 || !items[0].InlineStyle.SRTItalics || items[1].InlineStyle != nil || !items[2].InlineStyle.SRTBold {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestFormats(t *testing.T) {
	vtt := "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\n<i>a</i> <b>b</b>\n"
	if got := lineTexts(t, "a.vtt", vtt); len(got) != 1 || got[0] != "a b" {
		t.Errorf("vtt lines = %q", got)
	}
	if _, err := Read("a.docx", strings.NewReader("")); err == nil || !strings.Contains(err.Error(), "unsupported subtitle format") {
		t.Errorf("expected unsupported format error, got %v", err)
	}
}
