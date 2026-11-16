package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
)

// Style is how the subtitles look. Sizes and positions are in the coordinates of
// the Basic Title Motion template, which is authored at 1920x1080; Final Cut Pro
// scales the template to the project's resolution.
//
// Colors are "#RRGGBB", "#RRGGBBAA" or Final Cut Pro's "r g b a" with values 0-1.
// An empty color leaves that attribute out, so the template's own value is used.
type Style struct {
	Font     string  `json:"font"`
	FontFace string  `json:"fontFace"`
	FontSize float64 `json:"fontSize"`
	// FontColor is the text color.
	FontColor string `json:"fontColor"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	// StrokeColor and StrokeWidth draw an outline around the text.
	StrokeColor string  `json:"strokeColor"`
	StrokeWidth float64 `json:"strokeWidth"`
	// ShadowColor, ShadowDistance, ShadowAngle (degrees) and ShadowBlurRadius draw a drop shadow.
	ShadowColor      string  `json:"shadowColor"`
	ShadowDistance   float64 `json:"shadowDistance"`
	ShadowAngle      float64 `json:"shadowAngle"`
	ShadowBlurRadius float64 `json:"shadowBlurRadius"`
	LineSpacing      float64 `json:"lineSpacing"`
	// Alignment is "left", "center" or "right".
	Alignment string `json:"alignment"`
	// PositionX and PositionY place the text relative to the center of the frame.
	PositionX float64 `json:"positionX"`
	PositionY float64 `json:"positionY"`
}

// DefaultStyle is white PingFang SC with a soft shadow near the bottom of the frame.
func DefaultStyle() Style {
	return Style{
		Font:           "PingFang SC",
		FontFace:       "Semibold",
		FontSize:       52,
		FontColor:      "1 1 1 1",
		Bold:           true,
		ShadowColor:    "0 0 0 0.75",
		ShadowDistance: 5,
		ShadowAngle:    315,
		Alignment:      "center",
		PositionY:      -450,
	}
}

// LoadStyle reads a JSON style file. Fields missing from the file keep their
// value from base; unknown fields are an error, to catch typos.
func LoadStyle(path string, base Style) (Style, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return base, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	s := base
	if err := dec.Decode(&s); err != nil {
		return base, fmt.Errorf("style file %s: %w", path, err)
	}
	return s, nil
}

// Values of the Basic Title template's paragraph Alignment parameter (id 401).
var alignments = map[string]string{
	"left":   "0 (Left)",
	"center": "1 (Center)",
	"right":  "2 (Right)",
}

// Keys of the Basic Title template parameters that are set on every title.
const (
	positionKey  = "9999/999166631/999166633/1/100/101"
	alignmentKey = "9999/999166631/999166633/2/354/999169573/401"
	flattenKey   = "9999/999166631/999166633/2/351"
)

// textStyle validates the style and converts it to FCPXML text-style attributes
// and the title parameters that go with them.
func (s Style) textStyle() (fcpxml.TextStyle, []fcpxml.Param, error) {
	var ts fcpxml.TextStyle
	var err error
	fail := func(field string, e error) (fcpxml.TextStyle, []fcpxml.Param, error) {
		return fcpxml.TextStyle{}, nil, fmt.Errorf("style %s: %w", field, e)
	}

	if s.FontSize <= 0 {
		return fail("fontSize", fmt.Errorf("must be greater than 0, got %v", s.FontSize))
	}
	if s.StrokeWidth < 0 || s.ShadowBlurRadius < 0 || s.ShadowDistance < 0 {
		return fail("stroke/shadow", fmt.Errorf("widths, distances and radii cannot be negative"))
	}
	paramAlignment, ok := alignments[s.Alignment]
	if !ok {
		return fail("alignment", fmt.Errorf("must be left, center or right, got %q", s.Alignment))
	}
	if ts.FontColor, err = parseColor(s.FontColor); err != nil {
		return fail("fontColor", err)
	}
	if ts.StrokeColor, err = parseColor(s.StrokeColor); err != nil {
		return fail("strokeColor", err)
	}
	if ts.StrokeColor != "" && s.StrokeWidth == 0 {
		return fail("strokeWidth", fmt.Errorf("must be greater than 0 when strokeColor is set"))
	}
	if ts.ShadowColor, err = parseColor(s.ShadowColor); err != nil {
		return fail("shadowColor", err)
	}

	ts.Font = s.Font
	ts.FontFace = s.FontFace
	ts.FontSize = number(s.FontSize)
	ts.Bold = boolAttr(s.Bold)
	ts.Italic = boolAttr(s.Italic)
	if s.StrokeWidth > 0 {
		ts.StrokeWidth = number(s.StrokeWidth)
	}
	if ts.ShadowColor != "" {
		ts.ShadowOffset = number(s.ShadowDistance) + " " + number(s.ShadowAngle)
		if s.ShadowBlurRadius > 0 {
			ts.ShadowBlurRadius = number(s.ShadowBlurRadius)
		}
	}
	if s.LineSpacing != 0 {
		ts.LineSpacing = number(s.LineSpacing)
	}
	ts.Alignment = s.Alignment

	params := []fcpxml.Param{
		{Name: "Position", Key: positionKey, Value: number(s.PositionX) + " " + number(s.PositionY)},
		{Name: "Alignment", Key: alignmentKey, Value: paramAlignment},
		{Name: "Flatten", Key: flattenKey, Value: "1"},
	}
	return ts, params, nil
}

// parseColor converts "#RRGGBB", "#RRGGBBAA" or "r g b [a]" (0-1) to Final Cut
// Pro's "r g b a". An empty string stays empty.
func parseColor(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	var c []float64
	if hex, ok := strings.CutPrefix(s, "#"); ok {
		if len(hex) != 6 && len(hex) != 8 {
			return "", fmt.Errorf("%q: hex colors are #RRGGBB or #RRGGBBAA", s)
		}
		for i := 0; i < len(hex); i += 2 {
			v, err := strconv.ParseUint(hex[i:i+2], 16, 8)
			if err != nil {
				return "", fmt.Errorf("%q: invalid hex color", s)
			}
			c = append(c, math.Round(float64(v)/255*1e6)/1e6)
		}
	} else {
		for _, f := range strings.Fields(s) {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil || v < 0 || v > 1 {
				return "", fmt.Errorf("%q: components must be numbers from 0 to 1", s)
			}
			c = append(c, v)
		}
		if len(c) != 3 && len(c) != 4 {
			return "", fmt.Errorf("%q: expected \"r g b\" or \"r g b a\"", s)
		}
	}
	if len(c) == 3 {
		c = append(c, 1)
	}
	parts := make([]string, len(c))
	for i, v := range c {
		parts[i] = number(v)
	}
	return strings.Join(parts, " "), nil
}

func number(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func boolAttr(b bool) string {
	if b {
		return "1"
	}
	return ""
}
