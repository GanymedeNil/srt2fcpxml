package convert

import (
	"os"
	"path/filepath"
	"testing"
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
	ts, params, err := DefaultStyle().textStyle()
	if err != nil {
		t.Fatal(err)
	}
	if ts.Font != "PingFang SC" || ts.FontSize != "52" || ts.FontFace != "Semibold" ||
		ts.FontColor != "1 1 1 1" || ts.Bold != "1" || ts.ShadowColor != "0 0 0 0.75" ||
		ts.ShadowOffset != "5 315" || ts.Alignment != "center" {
		t.Errorf("default text style changed: %+v", ts)
	}
	if params[0].Value != "0 -450" || params[1].Value != "1 (Center)" || params[2].Value != "1" {
		t.Errorf("default title params changed: %+v", params)
	}
}

func TestStyleValidation(t *testing.T) {
	tests := map[string]func(*Style){
		"zero font size":        func(s *Style) { s.FontSize = 0 },
		"bad alignment":         func(s *Style) { s.Alignment = "middle" },
		"bad color":             func(s *Style) { s.FontColor = "white" },
		"stroke without width":  func(s *Style) { s.StrokeColor = "#000000" },
		"negative stroke width": func(s *Style) { s.StrokeWidth = -1 },
	}
	for name, mutate := range tests {
		s := DefaultStyle()
		mutate(&s)
		if _, _, err := s.textStyle(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestNoShadow(t *testing.T) {
	s := DefaultStyle()
	s.ShadowColor = ""
	ts, _, err := s.textStyle()
	if err != nil {
		t.Fatal(err)
	}
	if ts.ShadowColor != "" || ts.ShadowOffset != "" {
		t.Errorf("shadow attributes should be omitted: %+v", ts)
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
