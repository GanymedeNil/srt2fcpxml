package Gap

import (
	"fmt"
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap/Title"
	"srt2fcpxml/core/FcpXML/Resources"
	"srt2fcpxml/lib"
)

type Gap struct {
	Text     string         `xml:",chardata"`
	Name     string         `xml:"name,attr"`
	Offset   string         `xml:"offset,attr"`
	Duration string         `xml:"duration,attr"`
	Start    string         `xml:"start,attr"`
	Title    []*Title.Title `xml:"title"`
}

func NewGap(duration float64) *Gap {
	frameRate := Resources.GetFrameRate()
	frameRateR := Common.FrameDuration(frameRate)
	frameDurationMolecular, frameDurationDenominator, _ := Common.FrameDurationFormat(frameRate)
	return &Gap{
		Name:     "空隙",
		Offset:   "0s",
		Duration: fmt.Sprintf("%.f/%.fs", lib.Round(duration*frameRateR, 0)*frameDurationMolecular, frameDurationDenominator),
		Start:    fmt.Sprintf("%.f/%.fs", 3.6*frameDurationDenominator*frameDurationMolecular, frameDurationDenominator),
		Title:    nil,
	}
}

func (g *Gap) AddTitle(title *Title.Title) *Gap {
	g.Title = append(g.Title, title)
	return g
}
