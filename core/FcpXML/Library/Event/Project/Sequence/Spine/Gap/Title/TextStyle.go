package Title

import "fmt"

type TextStyleDef struct {
	Text      string `xml:",chardata"`
	ID        string `xml:"id,attr"`
	TextStyle struct {
		Text         string `xml:",chardata"`
		Font         string `xml:"font,attr"`
		FontSize     string `xml:"fontSize,attr"`
		FontFace     string `xml:"fontFace,attr"`
		FontColor    string `xml:"fontColor,attr"`
		Bold         string `xml:"bold,attr"`
		ShadowColor  string `xml:"shadowColor,attr"`
		ShadowOffset string `xml:"shadowOffset,attr"`
		Alignment    string `xml:"alignment,attr"`
	} `xml:"text-style"`
}

func NewTextStyleDef(index int) *TextStyleDef {
	return &TextStyleDef{
		ID: fmt.Sprintf("ts%d", index),
		TextStyle: struct {
			Text         string `xml:",chardata"`
			Font         string `xml:"font,attr"`
			FontSize     string `xml:"fontSize,attr"`
			FontFace     string `xml:"fontFace,attr"`
			FontColor    string `xml:"fontColor,attr"`
			Bold         string `xml:"bold,attr"`
			ShadowColor  string `xml:"shadowColor,attr"`
			ShadowOffset string `xml:"shadowOffset,attr"`
			Alignment    string `xml:"alignment,attr"`
		}{
			Font:         "PingFang SC",
			FontSize:     "52",
			FontFace:     "Semibold",
			FontColor:    "0.999993 1 1 1",
			Bold:         "1",
			ShadowColor:  "0 0 0 0.75",
			ShadowOffset: "5 315",
			Alignment:    "center",
		},
	}
}
