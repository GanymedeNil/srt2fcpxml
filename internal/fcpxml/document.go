// Package fcpxml models the subset of the FCPXML 1.7 format that srt2fcpxml writes:
// one library with one event and one project whose timeline is a gap with the
// subtitles anchored to it as titles.
//
// Element order and attributes follow FCPXMLv1_7.dtd as shipped with Final Cut Pro.
// None of the elements used here changed up to FCPXML 1.14.
package fcpxml

import (
	"encoding/xml"
)

// Version is the FCPXML version written. 1.7 is read by Final Cut Pro 10.4.6 and later.
const Version = "1.7"

const header = xml.Header + "<!DOCTYPE fcpxml>\n\n"

type Document struct {
	XMLName   xml.Name  `xml:"fcpxml"`
	Version   string    `xml:"version,attr"`
	Resources Resources `xml:"resources"`
	Library   Library   `xml:"library"`
}

// Marshal returns the complete document, including the XML declaration and DOCTYPE.
func (d *Document) Marshal() ([]byte, error) {
	body, err := xml.MarshalIndent(d, "", "    ")
	if err != nil {
		return nil, err
	}
	out := append([]byte(header), body...)
	return append(out, '\n'), nil
}

type Resources struct {
	Formats []Format `xml:"format"`
	Effects []Effect `xml:"effect"`
	Assets  []Asset  `xml:"asset"`
}

type Format struct {
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr,omitempty"`
	FrameDuration string `xml:"frameDuration,attr"`
	Width         int    `xml:"width,attr"`
	Height        int    `xml:"height,attr"`
	ColorSpace    string `xml:"colorSpace,attr,omitempty"`
}

type Effect struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	UID  string `xml:"uid,attr"`
}

// Asset is a media file. In FCPXML 1.7 its location is the src attribute; from
// 1.9 on it moved to a media-rep child element.
type Asset struct {
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	Src           string `xml:"src,attr"`
	Start         string `xml:"start,attr"`
	Duration      string `xml:"duration,attr"`
	HasVideo      string `xml:"hasVideo,attr,omitempty"`
	Format        string `xml:"format,attr,omitempty"`
	HasAudio      string `xml:"hasAudio,attr,omitempty"`
	AudioSources  string `xml:"audioSources,attr,omitempty"`
	AudioChannels string `xml:"audioChannels,attr,omitempty"`
	AudioRate     string `xml:"audioRate,attr,omitempty"`
}

type Library struct {
	Location string  `xml:"location,attr,omitempty"`
	Events   []Event `xml:"event"`
}

type Event struct {
	Name     string    `xml:"name,attr"`
	UID      string    `xml:"uid,attr"`
	Projects []Project `xml:"project"`
}

type Project struct {
	Name     string   `xml:"name,attr"`
	UID      string   `xml:"uid,attr"`
	ModDate  string   `xml:"modDate,attr"`
	Sequence Sequence `xml:"sequence"`
}

type Sequence struct {
	Duration    string `xml:"duration,attr"`
	Format      string `xml:"format,attr"`
	TCStart     string `xml:"tcStart,attr"`
	TCFormat    string `xml:"tcFormat,attr"`
	AudioLayout string `xml:"audioLayout,attr"`
	AudioRate   string `xml:"audioRate,attr"`
	Spine       Spine  `xml:"spine"`
}

// Spine holds either a gap or an asset-clip; srt2fcpxml never mixes them.
type Spine struct {
	Gaps       []Gap       `xml:"gap"`
	AssetClips []AssetClip `xml:"asset-clip"`
}

// AssetClip places a whole asset on the timeline, with titles anchored to it.
// Anchored titles' offsets are in the clip's local time, which begins at Start.
type AssetClip struct {
	Ref      string  `xml:"ref,attr"`
	Offset   string  `xml:"offset,attr"`
	Name     string  `xml:"name,attr"`
	Start    string  `xml:"start,attr"`
	Duration string  `xml:"duration,attr"`
	Format   string  `xml:"format,attr,omitempty"`
	TCFormat string  `xml:"tcFormat,attr,omitempty"`
	Titles   []Title `xml:"title"`
}

type Gap struct {
	Name     string  `xml:"name,attr"`
	Offset   string  `xml:"offset,attr"`
	Duration string  `xml:"duration,attr"`
	Start    string  `xml:"start,attr"`
	Titles   []Title `xml:"title"`
}

// Title is a title clip anchored to the gap. Its children must stay in DTD order:
// param*, text*, text-style-def*.
type Title struct {
	Name          string         `xml:"name,attr"`
	Lane          string         `xml:"lane,attr"`
	Offset        string         `xml:"offset,attr"`
	Ref           string         `xml:"ref,attr"`
	Duration      string         `xml:"duration,attr"`
	Start         string         `xml:"start,attr"`
	Params        []Param        `xml:"param"`
	Text          Text           `xml:"text"`
	TextStyleDefs []TextStyleDef `xml:"text-style-def"`
}

type Param struct {
	Name  string `xml:"name,attr"`
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

// Text holds runs of text, each styled by a text-style-def. Newlines in a run
// are written as &#xA; which Final Cut Pro shows as line breaks.
type Text struct {
	Runs []TextRun `xml:"text-style"`
}

type TextRun struct {
	Ref  string `xml:"ref,attr"`
	Text string `xml:",chardata"`
}

type TextStyleDef struct {
	ID    string    `xml:"id,attr"`
	Style TextStyle `xml:"text-style"`
}

// TextStyle holds the text-style attributes that apply to titles. Empty values
// are omitted, so the Motion template's own value is used.
type TextStyle struct {
	Font             string `xml:"font,attr,omitempty"`
	FontSize         string `xml:"fontSize,attr,omitempty"`
	FontFace         string `xml:"fontFace,attr,omitempty"`
	FontColor        string `xml:"fontColor,attr,omitempty"`
	Bold             string `xml:"bold,attr,omitempty"`
	Italic           string `xml:"italic,attr,omitempty"`
	StrokeColor      string `xml:"strokeColor,attr,omitempty"`
	StrokeWidth      string `xml:"strokeWidth,attr,omitempty"`
	ShadowColor      string `xml:"shadowColor,attr,omitempty"`
	ShadowOffset     string `xml:"shadowOffset,attr,omitempty"`
	ShadowBlurRadius string `xml:"shadowBlurRadius,attr,omitempty"`
	Alignment        string `xml:"alignment,attr,omitempty"`
	LineSpacing      string `xml:"lineSpacing,attr,omitempty"`
}
