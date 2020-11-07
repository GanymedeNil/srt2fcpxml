package Title

import "fmt"

type Text struct {
	Text      string `xml:",chardata"`
	TextStyle struct {
		Text string `xml:",chardata"`
		Ref  string `xml:"ref,attr"`
	} `xml:"text-style"`
}

func NewContent(index int, content string) *Text {
	return &Text{
		TextStyle: struct {
			Text string `xml:",chardata"`
			Ref  string `xml:"ref,attr"`
		}{
			Text: content,
			Ref:  fmt.Sprintf("ts%d", index),
		},
	}
}
