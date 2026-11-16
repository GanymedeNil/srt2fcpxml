package core

import (
	"encoding/xml"
	"errors"
	"fmt"
	"srt2fcpxml/core/FcpXML"
	"srt2fcpxml/core/FcpXML/Common"
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

// Srt2FcpXmlExport converts subtitles to an FCPXML document body (without the XML header).
func Srt2FcpXmlExport(projectName string, rate Common.Rate, subtitles *astisub.Subtitles, width, height int) ([]byte, error) {
	if rate.IsZero() {
		return nil, errors.New("frame rate must be set")
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resolution %dx%d", width, height)
	}
	if subtitles == nil || len(subtitles.Items) == 0 {
		return nil, errors.New("no subtitles found")
	}

	fcpxml := FcpXML.New()
	res := Resources.NewResources()
	res.SetEffect(Resources.NewEffect())
	format := Resources.NewFormat().
		SetWidth(width).
		SetHeight(height).
		SetFrameRate(rate).Render()
	res.SetFormat(format)
	fcpxml.SetResources(res)

	// Start and end frames are each rounded from absolute times (not start plus a
	// separately rounded duration), so subtitles that touch in the SRT share a
	// frame boundary instead of overlapping or leaving a 1-frame hole. A title
	// lasts at least one frame.
	var titles []*Title.Title
	var endFrames int64
	for index, item := range subtitles.Items {
		startFrame, endFrame := rate.Frames(item.StartAt), rate.Frames(item.EndAt)
		if endFrame <= startFrame {
			endFrame = startFrame + 1
		}
		if endFrame > endFrames {
			endFrames = endFrame
		}
		textStyleDef := Title.NewTextStyleDef(index + 1)
		text := Title.NewContent(index+1, func(lines []astisub.Line) string {
			var os []string
			for _, l := range lines {
				os = append(os, l.String())
			}
			return strings.Join(os, "\n")
		}(item.Lines))
		title := Title.NewTitle(rate, item.String(), startFrame, endFrame).SetTextStyleDef(textStyleDef).SetText(text)
		title.AddParam(Title.NewParams("Position", "9999/999166631/999166633/1/100/101", "0 -450"))
		title.AddParam(Title.NewParams("Alignment", "9999/999166631/999166633/2/354/999169573/401", "1 (Center)"))
		title.AddParam(Title.NewParams("Flatten", "9999/999166631/999166633/2/351", "1"))
		titles = append(titles, title)
	}

	gap := Gap.NewGap(rate, endFrames)
	for _, title := range titles {
		gap.AddTitle(title)
	}

	spine := Spine.NewSpine().SetGap(gap)
	seq := Sequence.NewSequence(rate, endFrames).SetSpine(spine)
	project := Project.NewProject(projectName).SetSequence(seq)
	event := Event.NewEvent().SetProject(project)
	library := Library.NewLibrary(projectName).SetEvent(event)
	fcpxml.SetLibrary(library)

	return xml.MarshalIndent(fcpxml, "", "    ")
}
