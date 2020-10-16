package core

import (
	"encoding/xml"
	"fmt"
	"github.com/asticode/go-astisub"
	"os"
	"srt2fcpxml/lib"
	"strconv"
	"strings"
	"time"
)

type Fcpxml struct {
	XMLName   xml.Name `xml:"fcpxml"`
	Text      string   `xml:",chardata"`
	Version   string   `xml:"version,attr"`
	Resources struct {
		Text   string `xml:",chardata"`
		Format struct {
			Text          string `xml:",chardata"`
			ID            string `xml:"id,attr"`
			Name          string `xml:"name,attr"`
			FrameDuration string `xml:"frameDuration,attr"`
			Width         string `xml:"width,attr"`
			Height        string `xml:"height,attr"`
			ColorSpace    string `xml:"colorSpace,attr"`
		} `xml:"format"`
		Effect struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Name string `xml:"name,attr"`
			Uid  string `xml:"uid,attr"`
		} `xml:"effect"`
	} `xml:"resources"`
	Library struct {
		Text     string `xml:",chardata"`
		Location string `xml:"location,attr"`
		Event    struct {
			Text    string `xml:",chardata"`
			Name    string `xml:"name,attr"`
			Uid     string `xml:"uid,attr"`
			Project struct {
				Text     string `xml:",chardata"`
				Name     string `xml:"name,attr"`
				Uid      string `xml:"uid,attr"`
				ModDate  string `xml:"modDate,attr"`
				Sequence struct {
					Text        string `xml:",chardata"`
					Duration    string `xml:"duration,attr"`
					Format      string `xml:"format,attr"`
					TcStart     string `xml:"tcStart,attr"`
					TcFormat    string `xml:"tcFormat,attr"`
					AudioLayout string `xml:"audioLayout,attr"`
					AudioRate   string `xml:"audioRate,attr"`
					Spine       struct {
						Text string `xml:",chardata"`
						Gap  struct {
							Text     string  `xml:",chardata"`
							Name     string  `xml:"name,attr"`
							Offset   string  `xml:"offset,attr"`
							Duration string  `xml:"duration,attr"`
							Start    string  `xml:"start,attr"`
							Title    []title `xml:"title"`
						} `xml:"gap"`
					} `xml:"spine"`
				} `xml:"sequence"`
			} `xml:"project"`
		} `xml:"event"`
	} `xml:"library"`
}

type format struct {
	Text          string `xml:",chardata"`
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	FrameDuration string `xml:"frameDuration,attr"`
	Width         string `xml:"width,attr"`
	Height        string `xml:"height,attr"`
	ColorSpace    string `xml:"colorSpace,attr"`
}
type effect struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Uid  string `xml:"uid,attr"`
}
type title struct {
	Chardata string `xml:",chardata"`
	Name     string `xml:"name,attr"`
	Lane     string `xml:"lane,attr"`
	Offset   string `xml:"offset,attr"`
	Ref      string `xml:"ref,attr"`
	Duration string `xml:"duration,attr"`
	Start    string `xml:"start,attr"`
	Param    []struct {
		Text  string `xml:",chardata"`
		Name  string `xml:"name,attr"`
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	} `xml:"param"`
	Text struct {
		Text      string `xml:",chardata"`
		TextStyle struct {
			Text string `xml:",chardata"`
			Ref  string `xml:"ref,attr"`
		} `xml:"text-style"`
	} `xml:"text"`
	TextStyleDef struct {
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
	} `xml:"text-style-def"`
}

var param = []struct {
	Text  string `xml:",chardata"`
	Name  string `xml:"name,attr"`
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}{
	{
		Name:  "位置",
		Key:   "9999/999166631/999166633/1/100/101",
		Value: "0 -450",
	}, {
		Name:  "对齐",
		Key:   "9999/999166631/999166633/2/354/999169573/401",
		Value: "1 (居中)",
	}, {
		Name:  "展平",
		Key:   "9999/999166631/999166633/2/351",
		Value: "1",
	},
}

const (
	startTime = 3600
)

var (
	frameDuration int
)

func Srt2FcpxmlExport(projectName string, fd int, subtitles *astisub.Subtitles) ([]byte, error) {
	frameDuration = fd
	result := Fcpxml{}
	result.Version = "1.7"
	result.Resources.Format = format{
		ID:            "r1",
		Name:          fmt.Sprintf("%s%dp%d", "FFVideoFormat", 1080, frameDuration),
		FrameDuration: fmt.Sprintf("%d/%ds", 100, 100*frameDuration),
		Width:         "1920",
		Height:        "1080",
		ColorSpace:    "1-1-1 (Rec. 709)",
	}
	result.Resources.Effect = effect{
		ID:   "r2",
		Name: "Basic Title",
		Uid:  ".../Titles.localized/Bumper:Opener.localized/Basic Title.localized/Basic Title.moti",
	}
	homeDir, _ := os.UserHomeDir()
	result.Library.Location = fmt.Sprintf("file://%s/Movies/%s.fcpbundle", homeDir, projectName)

	nowTime := time.Now()
	result.Library.Event.Name = nowTime.Format("2006-01-02")
	result.Library.Event.Uid = "425137E1-800B-4C9D-B029-07697EE6F56A"

	result.Library.Event.Project.Name = projectName
	result.Library.Event.Project.Uid = "93D8C871-F60A-4DF7-BFF4-D2FFB8C554A9"
	result.Library.Event.Project.ModDate = nowTime.Format("2006-01-02 15:04:05 +0800")

	result.Library.Event.Project.Sequence.Duration = fmt.Sprintf("%vs", lib.Round(subtitles.Duration().Seconds(), 0))
	result.Library.Event.Project.Sequence.Format = "r1"
	result.Library.Event.Project.Sequence.TcStart = "0s"
	result.Library.Event.Project.Sequence.TcFormat = "NDF"
	result.Library.Event.Project.Sequence.AudioLayout = "stereo"
	result.Library.Event.Project.Sequence.AudioRate = "48k"

	result.Library.Event.Project.Sequence.Spine.Gap.Name = "空隙"
	result.Library.Event.Project.Sequence.Spine.Gap.Offset = "0s"
	result.Library.Event.Project.Sequence.Spine.Gap.Duration = fmt.Sprintf("%vs", lib.Round(subtitles.Duration().Seconds(), 0))
	result.Library.Event.Project.Sequence.Spine.Gap.Start = fmt.Sprintf("%ds", startTime)
	result.Library.Event.Project.Sequence.Spine.Gap.Title = texts(subtitles.Items)
	return xml.MarshalIndent(result, "", "    ")
}

func texts(subtitles []*astisub.Item) []title {
	var stitles []title
	for index, item := range subtitles {
		title := title{
			Name: item.String(),
			Lane: "1",
			Offset: fmt.Sprintf("%.f/%vs",
				lib.Round(item.StartAt.Seconds()*float64(frameDuration)*100, -2)+
					(float64(frameDuration)*100*startTime), float64(frameDuration)*100),
			Ref: "r2",
			Duration: fmt.Sprintf("%.f/%vs",
				lib.Round((item.EndAt.Seconds()-item.StartAt.Seconds())*float64(frameDuration)*100, -2),
				float64(frameDuration)*100),
			Start: strconv.Itoa(startTime),
			Param: param,
			Text: struct {
				Text      string `xml:",chardata"`
				TextStyle struct {
					Text string `xml:",chardata"`
					Ref  string `xml:"ref,attr"`
				} `xml:"text-style"`
			}{TextStyle: struct {
				Text string `xml:",chardata"`
				Ref  string `xml:"ref,attr"`
			}{Text: func(lines []astisub.Line) string {
				var os []string
				for _, l := range lines {
					os = append(os, l.String())
				}
				return strings.Join(os, "\n")
			}(item.Lines), Ref: fmt.Sprintf("ts%d", index+1)}},
			TextStyleDef: struct {
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
			}{
				ID: fmt.Sprintf("ts%d", index+1),
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
			},
		}
		stitles = append(stitles, title)
	}
	return stitles
}
