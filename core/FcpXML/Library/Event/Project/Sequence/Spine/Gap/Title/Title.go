package Title

import (
	"fmt"
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Resources"
	"srt2fcpxml/lib"
)

type Title struct {
	Chardata     string        `xml:",chardata"`
	Name         string        `xml:"name,attr"`
	Lane         string        `xml:"lane,attr"`
	Offset       string        `xml:"offset,attr"`
	Ref          string        `xml:"ref,attr"`
	Duration     string        `xml:"duration,attr"`
	Start        string        `xml:"start,attr"`
	Param        []*Param      `xml:"param"`
	Text         *Text         `xml:"text"`
	TextStyleDef *TextStyleDef `xml:"text-style-def"`
}

func NewTitle(name string, start, end float64) *Title {
	frameRate := Resources.GetFrameRate()
	frameRateR := Common.FrameDuration(frameRate)
	frameDurationMolecular, frameDurationDenominator, _ := Common.FrameDurationFormat(frameRate)
	projectStart := 3.6 * frameDurationDenominator * frameDurationMolecular
	return &Title{
		Name:         name,
		Lane:         "1",
		Offset:       fmt.Sprintf("%.f/%.fs", lib.Round(start*frameRateR, 0)*frameDurationMolecular+projectStart, frameDurationDenominator),
		Ref:          "r2",
		Duration:     fmt.Sprintf("%.f/%.fs", lib.Round((end-start)*frameRateR, 0)*frameDurationMolecular*120000.0/frameDurationDenominator, 120000.0),
		Start:        fmt.Sprintf("%.f/%.fs", projectStart, frameDurationDenominator),
		Param:        []*Param{},
		Text:         &Text{},
		TextStyleDef: &TextStyleDef{},
	}
}

func (t *Title) AddParam(param *Param) *Title {
	t.Param = append(t.Param, param)
	return t
}

func (t *Title) SetText(text *Text) *Title {
	t.Text = text
	return t
}

func (t *Title) SetTextStyleDef(textStyleDef *TextStyleDef) *Title {
	t.TextStyleDef = textStyleDef
	return t
}
