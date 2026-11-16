package FcpXML

import (
	"encoding/xml"
	"srt2fcpxml/core/FcpXML/Common"
	"srt2fcpxml/core/FcpXML/Library"
	"srt2fcpxml/core/FcpXML/Library/Event"
	"srt2fcpxml/core/FcpXML/Library/Event/Project"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap"
	"srt2fcpxml/core/FcpXML/Library/Event/Project/Sequence/Spine/Gap/Title"
	"srt2fcpxml/core/FcpXML/Resources"
	"testing"
	"time"
)

func TestFcpXML_SetVersion(t *testing.T) {
	rate, err := Common.ParseRate("23.98")
	if err != nil {
		t.Fatal(err)
	}
	fcpxml := New()
	res := Resources.NewResources()
	res.SetEffect(Resources.NewEffect())
	format := Resources.NewFormat().
		SetWidth(1440).
		SetHeight(1080).
		SetFrameRate(rate).Render()
	res.SetFormat(format)
	fcpxml.SetResources(res)

	gap := Gap.NewGap(rate, rate.Frames(180200*time.Millisecond))

	textStyleDef := Title.NewTextStyleDef(1)
	text := Title.NewContent(1, "4:00")
	title := Title.NewTitle(rate, "4:00 - 基本字幕", rate.Frames(4*time.Second), rate.Frames(7800*time.Millisecond)).SetTextStyleDef(textStyleDef).SetText(text)
	title.AddParam(Title.NewParams("位置", "9999/999166631/999166633/1/100/101", "0 -450"))
	title.AddParam(Title.NewParams("对齐", "9999/999166631/999166633/2/354/999169573/401", "1 (居中)"))
	title.AddParam(Title.NewParams("展平", "9999/999166631/999166633/2/351", "1"))
	gap.AddTitle(title)

	spine := Spine.NewSpine().SetGap(gap)

	seq := Sequence.NewSequence(rate, rate.Frames(180200*time.Millisecond)).SetSpine(spine)

	project := Project.NewProject("ceshi").SetSequence(seq)

	event := Event.NewEvent().SetProject(project)

	lib := Library.NewLibrary("未命名 1").SetEvent(event)

	fcpxml.SetLibrary(lib)

	e, _ := xml.MarshalIndent(&fcpxml, "", "    ")
	t.Log(string(e))
}
