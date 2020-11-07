package Project

import (
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Project struct {
	Text     string             `xml:",chardata"`
	Name     string             `xml:"name,attr"`
	Uid      string             `xml:"uid,attr"`
	ModDate  string             `xml:"modDate,attr"`
	Sequence *Sequence.Sequence `xml:"sequence"`
}

func NewProject(projectName string) *Project {
	modDate := time.Now().Format("2006-01-02 15:04:05 -0700")
	return &Project{
		Name:     projectName,
		Uid:      uuid.NewV4().String(),
		ModDate:  modDate,
		Sequence: &Sequence.Sequence{},
	}
}

func (p *Project) SetSequence(sequence *Sequence.Sequence) *Project {
	p.Sequence = sequence
	return p
}
