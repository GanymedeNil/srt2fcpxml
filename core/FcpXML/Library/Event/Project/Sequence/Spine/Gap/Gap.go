package Gap

import (
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap/Title"
)

type Gap struct {
	Text     string         `xml:",chardata"`
	Name     string         `xml:"name,attr"`
	Offset   string         `xml:"offset,attr"`
	Duration string         `xml:"duration,attr"`
	Start    string         `xml:"start,attr"`
	Title    []*Title.Title `xml:"title"`
}

// NewGap creates the gap the titles are anchored to; durationFrames is its length in frames.
func NewGap(rate Common.Rate, durationFrames int64) *Gap {
	return &Gap{
		Name:     "空隙",
		Offset:   "0s",
		Duration: rate.Time(durationFrames),
		Start:    rate.Time(rate.GapStartFrames()),
		Title:    nil,
	}
}

func (g *Gap) AddTitle(title *Title.Title) *Gap {
	g.Title = append(g.Title, title)
	return g
}
