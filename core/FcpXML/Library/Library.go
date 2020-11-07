package Library

import (
	"fmt"
	"net/url"
	"os"
	"srt2fcpxml/core/FcpXML/Library/Event"
)

type Library struct {
	Text     string       `xml:",chardata"`
	Location string       `xml:"location,attr"`
	Event    *Event.Event `xml:"event"`
}

func NewLibrary(fileName string) *Library {
	homeDir, _ := os.UserHomeDir()
	location := fmt.Sprintf("file://%s/Movies/%s.fcpbundle", homeDir, url.PathEscape(fileName))
	return &Library{
		Text:     "",
		Location: location,
		Event:    &Event.Event{},
	}
}

func (l *Library) SetEvent(event *Event.Event) *Library {
	l.Event = event
	return l
}
