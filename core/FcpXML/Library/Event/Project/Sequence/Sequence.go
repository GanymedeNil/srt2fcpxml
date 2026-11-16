package Sequence

import (
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine"
)

type Sequence struct {
	Text        string       `xml:",chardata"`
	Duration    string       `xml:"duration,attr"`
	Format      string       `xml:"format,attr"`
	TcStart     string       `xml:"tcStart,attr"`
	TcFormat    string       `xml:"tcFormat,attr"`
	AudioLayout string       `xml:"audioLayout,attr"`
	AudioRate   string       `xml:"audioRate,attr"`
	Spine       *Spine.Spine `xml:"spine"`
}

// NewSequence creates the project sequence; durationFrames is its length in frames.
func NewSequence(rate Common.Rate, durationFrames int64) *Sequence {
	return &Sequence{
		Text:        "",
		Duration:    rate.Time(durationFrames),
		Format:      "r1",
		TcStart:     "0s",
		TcFormat:    "NDF",
		AudioLayout: "stereo",
		AudioRate:   "48k",
		Spine:       &Spine.Spine{},
	}
}

func (s *Sequence) SetSpine(spine *Spine.Spine) *Sequence {
	s.Spine = spine
	return s
}
