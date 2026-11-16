package Title

import (
	"srt2fcpxml/core/FcpXML/Common"
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

// NewTitle creates a title anchored to the gap. startFrame and endFrame are
// positions on the timeline in whole frames.
func NewTitle(rate Common.Rate, name string, startFrame, endFrame int64) *Title {
	base := rate.GapStartFrames()
	return &Title{
		Name:         name,
		Lane:         "1",
		Offset:       rate.Time(base + startFrame),
		Ref:          "r2",
		Duration:     rate.Time(endFrame - startFrame),
		Start:        rate.Time(base),
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
