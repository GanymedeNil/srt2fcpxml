package core

import (
	"encoding/xml"
	"srt2fcpxml/core/FcpXML"
	"srt2fcpxml/core/FcpXML/Library"
	"srt2fcpxml/core/FcpXML/Library/Event"
	"srt2fcpxml/core/FcpXML/Library/Event/Project"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap/Title"
	"srt2fcpxml/core/FcpXML/Resources"
	"strings"

	"github.com/asticode/go-astisub"
)

func Srt2FcpXmlExport(projectName string, frameDuration interface{}, subtitles *astisub.Subtitles) ([]byte, error) {
	fcpxml := FcpXML.New()
	res := Resources.NewResources()
	res.SetEffect(Resources.NewEffect())
	format := Resources.NewFormat().
		SetWidth(1920).
		SetHeight(1080).
		SetFrameRate(frameDuration).Render()
	res.SetFormat(format)
	fcpxml.SetResources(res)
	gap := Gap.NewGap(subtitles.Duration().Seconds())

	for index, item := range subtitles.Items {
		textStyleDef := Title.NewTextStyleDef(index + 1)
		text := Title.NewContent(index+1, func(lines []astisub.Line) string {
			var os []string
			for _, l := range lines {
				os = append(os, l.String())
			}
			return strings.Join(os, "\n")
		}(item.Lines))
		title := Title.NewTitle(item.String(), item.StartAt.Seconds(), item.EndAt.Seconds()).SetTextStyleDef(textStyleDef).SetText(text)
		title.AddParam(Title.NewParams("位置", "9999/999166631/999166633/1/100/101", "0 -450"))
		title.AddParam(Title.NewParams("对齐", "9999/999166631/999166633/2/354/999169573/401", "1 (居中)"))
		title.AddParam(Title.NewParams("展平", "9999/999166631/999166633/2/351", "1"))
		gap.AddTitle(title)
	}

	spine := Spine.NewSpine().SetGap(gap)
	seq := Sequence.NewSequence(subtitles.Duration().Seconds()).SetSpine(spine)
	project := Project.NewProject(projectName).SetSequence(seq)
	event := Event.NewEvent().SetProject(project)
	library := Library.NewLibrary(projectName).SetEvent(event)
	fcpxml.SetLibrary(library)

	return xml.MarshalIndent(fcpxml, "", "    ")
}
