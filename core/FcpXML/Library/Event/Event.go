package Event

import (
	"srt2fcpxml/core/FcpXML/Library/Event/Project"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Event struct {
	Text    string           `xml:",chardata"`
	Name    string           `xml:"name,attr"`
	Uid     string           `xml:"uid,attr"`
	Project *Project.Project `xml:"project"`
}

func NewEvent() *Event {
	return &Event{
		Name:    time.Now().Format("2006-01-02"),
		Uid:     uuid.NewV4().String(),
		Project: &Project.Project{},
	}
}

func (e *Event) SetProject(project *Project.Project) *Event {
	e.Project = project
	return e
}
