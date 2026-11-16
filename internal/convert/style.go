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
	"github.com/asticode/go-astisub"
)

// TextLook is how a run of text looks.
//
// Colors are "#RRGGBB", "#RRGGBBAA" or Final Cut Pro's "r g b a" with values 0-1.
// An empty color leaves that attribute out, so the template's own value is used.
type TextLook struct {
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
}

// Style is how the subtitles look. Sizes and positions are in the coordinates of
// the Basic Title Motion template, which is authored at 1920x1080; Final Cut Pro
// scales the template to the project's resolution.
//
// The embedded TextLook styles the primary language. In JSON its fields sit at
// the top level, next to the layout fields; Secondary is a nested object.
type Style struct {
	TextLook
	LineSpacing float64 `json:"lineSpacing"`
	// Alignment is "left", "center" or "right".
	Alignment string `json:"alignment"`
	// PositionX and PositionY place the text relative to the center of the frame.
	PositionX float64 `json:"positionX"`
	PositionY float64 `json:"positionY"`
	// Secondary styles the second language of bilingual subtitles.
	Secondary TextLook `json:"secondary"`
}

// DefaultStyle is white PingFang SC with a soft shadow near the bottom of the
// frame, with a smaller regular-weight second language.
func DefaultStyle() Style {
	shadow := TextLook{ShadowColor: "0 0 0 0.75", ShadowDistance: 5, ShadowAngle: 315}
	primary, secondary := shadow, shadow
	primary.Font, primary.FontFace, primary.FontSize, primary.FontColor, primary.Bold = "PingFang SC", "Semibold", 52, "1 1 1 1", true
	secondary.Font, secondary.FontFace, secondary.FontSize, secondary.FontColor = "PingFang SC", "Regular", 36, "1 1 1 1"
	return Style{
		TextLook:  primary,
		Alignment: "center",
		PositionY: -450,
		Secondary: secondary,
	}
}

// LoadStyle reads a JSON style file. Fields missing from the file keep their
// value from base; unknown fields are an error, to catch typos.
func LoadStyle(path string, base Style) (Style, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return base, err
	}
	s, err := ParseStyle(data, base)
	if err != nil {
		return base, fmt.Errorf("style file %s: %w", path, err)
	}
	return s, nil
}

// ParseStyle is LoadStyle for JSON that is already in memory.
func ParseStyle(data []byte, base Style) (Style, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	s := base
	if err := dec.Decode(&s); err != nil {
		return base, err
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

// compiledStyle is a validated Style in FCPXML terms.
type compiledStyle struct {
	primary, secondary fcpxml.TextStyle
	params             []fcpxml.Param
}

func (s Style) compile() (compiledStyle, error) {
	paramAlignment, ok := alignments[s.Alignment]
	if !ok {
		return compiledStyle{}, fmt.Errorf("style alignment: must be left, center or right, got %q", s.Alignment)
	}
	primary, err := s.TextLook.textStyle("")
	if err != nil {
		return compiledStyle{}, err
	}
	secondary, err := s.Secondary.textStyle("secondary ")
	if err != nil {
		return compiledStyle{}, err
	}
	for _, ts := range []*fcpxml.TextStyle{&primary, &secondary} {
		ts.Alignment = s.Alignment
		if s.LineSpacing != 0 {
			ts.LineSpacing = number(s.LineSpacing)
		}
	}
	return compiledStyle{
		primary:   primary,
		secondary: secondary,
		params: []fcpxml.Param{
			{Name: "Position", Key: positionKey, Value: number(s.PositionX) + " " + number(s.PositionY)},
			{Name: "Alignment", Key: alignmentKey, Value: paramAlignment},
			{Name: "Flatten", Key: flattenKey, Value: "1"},
		},
	}, nil
}

// textStyle validates the look and converts it to FCPXML text-style attributes.
// prefix names the look in error messages.
func (l TextLook) textStyle(prefix string) (fcpxml.TextStyle, error) {
	var ts fcpxml.TextStyle
	var err error
	fail := func(field string, e error) (fcpxml.TextStyle, error) {
		return fcpxml.TextStyle{}, fmt.Errorf("style %s%s: %w", prefix, field, e)
	}

	if l.FontSize <= 0 {
		return fail("fontSize", fmt.Errorf("must be greater than 0, got %v", l.FontSize))
	}
	if l.StrokeWidth < 0 || l.ShadowBlurRadius < 0 || l.ShadowDistance < 0 {
		return fail("stroke/shadow", fmt.Errorf("widths, distances and radii cannot be negative"))
	}
	if ts.FontColor, err = parseColor(l.FontColor); err != nil {
		return fail("fontColor", err)
	}
	if ts.StrokeColor, err = parseColor(l.StrokeColor); err != nil {
		return fail("strokeColor", err)
	}
	if ts.StrokeColor != "" && l.StrokeWidth == 0 {
		return fail("strokeWidth", fmt.Errorf("must be greater than 0 when strokeColor is set"))
	}
	if ts.ShadowColor, err = parseColor(l.ShadowColor); err != nil {
		return fail("shadowColor", err)
	}

	ts.Font = l.Font
	ts.FontFace = l.FontFace
	ts.FontSize = number(l.FontSize)
	ts.Bold = boolAttr(l.Bold)
	ts.Italic = boolAttr(l.Italic)
	if l.StrokeWidth > 0 {
		ts.StrokeWidth = number(l.StrokeWidth)
	}
	if ts.ShadowColor != "" {
		ts.ShadowOffset = number(l.ShadowDistance) + " " + number(l.ShadowAngle)
		if l.ShadowBlurRadius > 0 {
			ts.ShadowBlurRadius = number(l.ShadowBlurRadius)
		}
	}
	return ts, nil
}

// inline is the styling a subtitle file applies to part of a line.
type inline struct {
	bold, italic bool
	color        string // Final Cut Pro "r g b a", or "" for none
}

// inlineStyle reads bold, italic (SRT, WebVTT, TTML, SSA) and color (SRT) from a
// line item. Underline is not supported by text-style in FCPXML 1.7.
func inlineStyle(a *astisub.StyleAttributes) inline {
	var in inline
	if a == nil {
		return in
	}
	in.bold = a.SRTBold || (a.SSABold != nil && *a.SSABold) ||
		(a.TTMLFontWeight != nil && *a.TTMLFontWeight == "bold")
	in.italic = a.SRTItalics || (a.SSAItalic != nil && *a.SSAItalic) ||
		(a.TTMLFontStyle != nil && *a.TTMLFontStyle == "italic")
	for _, tag := range a.WebVTTTags {
		switch tag.Name {
		case "b":
			in.bold = true
		case "i":
			in.italic = true
		}
	}
	if a.SRTColor != nil {
		// Colors astisub cannot decode (e.g. rgb()) are left out.
		in.color, _ = parseColor(a.SRTColor.HTMLString())
	}
	return in
}

// apply returns ts with the inline styling on top.
func (in inline) apply(ts fcpxml.TextStyle) fcpxml.TextStyle {
	if in.bold {
		ts.Bold = "1"
	}
	if in.italic {
		ts.Italic = "1"
	}
	if in.color != "" {
		ts.FontColor = in.color
	}
	return ts
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
