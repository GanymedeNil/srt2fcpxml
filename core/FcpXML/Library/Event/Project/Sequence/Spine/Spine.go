package Spine

import "srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap"

type Spine struct {
	Text string   `xml:",chardata"`
	Gap  *Gap.Gap `xml:"gap"`
}

func NewSpine() *Spine {
	return &Spine{
		Gap: &Gap.Gap{},
	}
}
func (s *Spine) SetGap(gap *Gap.Gap) *Spine {
	s.Gap = gap
	return s
}
