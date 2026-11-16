package convert

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/asticode/go-astisub"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{in: "", want: ""},
		{in: "#FFFFFF", want: "1 1 1 1"},
		{in: "#000000cc", want: "0 0 0 0.8"},
		{in: "#FFD700", want: "1 0.843137 0 1"},
		{in: "0.999993 1 1 1", want: "0.999993 1 1 1"},
		{in: "0 0 0", want: "0 0 0 1"},
		{in: " 0 0 0 0.75 ", want: "0 0 0 0.75"},
		{in: "#FFF", wantErr: true},
		{in: "#GG0000", wantErr: true},
		{in: "1 1", wantErr: true},
		{in: "1 1 1 1 1", wantErr: true},
		{in: "255 0 0", wantErr: true},
		{in: "red", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseColor(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseColor(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("parseColor(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDefaultStyle(t *testing.T) {
	c, err := DefaultStyle().compile()
	if err != nil {
		t.Fatal(err)
	}
	ts := c.primary
	if ts.Font != "PingFang SC" || ts.FontSize != "52" || ts.FontFace != "Semibold" ||
		ts.FontColor != "1 1 1 1" || ts.Bold != "1" || ts.ShadowColor != "0 0 0 0.75" ||
		ts.ShadowOffset != "5 315" || ts.Alignment != "center" {
		t.Errorf("default text style changed: %+v", ts)
	}
	sec := c.secondary
	if sec.Font != "PingFang SC" || sec.FontSize != "36" || sec.FontFace != "Regular" || sec.Bold != "" ||
		sec.ShadowOffset != "5 315" || sec.Alignment != "center" {
		t.Errorf("default secondary style changed: %+v", sec)
	}
	if p := c.params; p[0].Value != "0 -450" || p[1].Value != "1 (Center)" || p[2].Value != "1" {
		t.Errorf("default title params changed: %+v", p)
	}
}

func TestStyleValidation(t *testing.T) {
	tests := map[string]func(*Style){
		"zero font size":           func(s *Style) { s.FontSize = 0 },
		"bad alignment":            func(s *Style) { s.Alignment = "middle" },
		"bad color":                func(s *Style) { s.FontColor = "white" },
		"stroke without width":     func(s *Style) { s.StrokeColor = "#000000" },
		"negative stroke width":    func(s *Style) { s.StrokeWidth = -1 },
		"bad secondary font size":  func(s *Style) { s.Secondary.FontSize = -2 },
		"bad secondary font color": func(s *Style) { s.Secondary.FontColor = "#12" },
	}
	for name, mutate := range tests {
		s := DefaultStyle()
		mutate(&s)
		if _, err := s.compile(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestNoShadow(t *testing.T) {
	s := DefaultStyle()
	s.ShadowColor = ""
	c, err := s.compile()
	if err != nil {
		t.Fatal(err)
	}
	if c.primary.ShadowColor != "" || c.primary.ShadowOffset != "" {
		t.Errorf("shadow attributes should be omitted: %+v", c.primary)
	}
}

func TestInlineStyle(t *testing.T) {
	base := fcpxml.TextStyle{FontColor: "1 1 1 1"}
	yes := true
	weight := "bold"
	tests := []struct {
		name string
		in   *astisub.StyleAttributes
		want fcpxml.TextStyle
	}{
		{"none", nil, base},
		{"srt italic", &astisub.StyleAttributes{SRTItalics: true}, fcpxml.TextStyle{FontColor: "1 1 1 1", Italic: "1"}},
		{"srt color", &astisub.StyleAttributes{SRTColor: &astisub.Color{Red: 255}}, fcpxml.TextStyle{FontColor: "1 0 0 1"}},
		{"webvtt bold", &astisub.StyleAttributes{WebVTTTags: []astisub.WebVTTTag{{Name: "b"}}}, fcpxml.TextStyle{FontColor: "1 1 1 1", Bold: "1"}},
		{"ssa italic", &astisub.StyleAttributes{SSAItalic: &yes}, fcpxml.TextStyle{FontColor: "1 1 1 1", Italic: "1"}},
		{"ttml bold", &astisub.StyleAttributes{TTMLFontWeight: &weight}, fcpxml.TextStyle{FontColor: "1 1 1 1", Bold: "1"}},
	}
	for _, tt := range tests {
		if got := inlineStyle(tt.in).apply(base); got != tt.want {
			t.Errorf("%s: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestLoadStyle(t *testing.T) {
	dir := t.TempDir()
	partial := filepath.Join(dir, "partial.json")
	os.WriteFile(partial, []byte(`{"fontSize": 60, "bold": false}`), 0o644)
	s, err := LoadStyle(partial, DefaultStyle())
	if err != nil {
		t.Fatal(err)
	}
	if s.FontSize != 60 || s.Bold || s.Font != "PingFang SC" || s.PositionY != -450 {
		t.Errorf("fields in the file should override, others keep defaults: %+v", s)
	}

	typo := filepath.Join(dir, "typo.json")
	os.WriteFile(typo, []byte(`{"fontSzie": 60}`), 0o644)
	if _, err := LoadStyle(typo, DefaultStyle()); err == nil {
		t.Error("unknown field should be an error")
	}
	if _, err := LoadStyle(filepath.Join(dir, "missing.json"), DefaultStyle()); err == nil {
		t.Error("missing file should be an error")
	}
}
