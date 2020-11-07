package Sequence

import (
	"fmt"
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine"
	"srt2fcpxml/core/FcpXML/Resources"
	"srt2fcpxml/lib"
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

func NewSequence(duration float64) *Sequence {
	frameRate := Resources.GetFrameRate()
	frameRateR := Common.FrameDuration(frameRate)
	frameDurationMolecular, frameDurationDenominator, _ := Common.FrameDurationFormat(frameRate)
	return &Sequence{
		Text:        "",
		Duration:    fmt.Sprintf("%.f/%.fs", lib.Round(duration*frameRateR, 0)*frameDurationMolecular, frameDurationDenominator),
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
