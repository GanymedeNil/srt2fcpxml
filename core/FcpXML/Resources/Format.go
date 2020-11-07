package Resources

import (
	"fmt"
	"srt2fcpxml/core/FcpXML/Common"
)

type format struct {
	Text                 string      `xml:",chardata"`
	ID                   string      `xml:"id,attr"`
	Name                 string      `xml:"name,attr"`
	FrameRate            interface{} `xml:"-"`
	FrameDuration        string      `xml:"frameDuration,attr"`
	FrameDurationCompute float64     `xml:"-"`
	Width                int         `xml:"width,attr"`
	Height               int         `xml:"height,attr"`
	ColorSpace           string      `xml:"colorSpace,attr"`
}
type RenderFormat format

var render RenderFormat

func NewFormat() *format {
	render.ID = "r1"
	render.Width = 1920
	render.Height = 1080
	render.ColorSpace = "1-1-1 (Rec. 709)"
	return &format{}
}

func (f *format) SetWidth(width int) *format {
	render.Width = width
	return f
}
func (f *format) SetHeight(height int) *format {
	render.Height = height
	return f
}
func (f *format) SetFrameRate(frameRate interface{}) *format {
	render.FrameRate = frameRate
	return f
}

func (f *format) SetColorSpace(colorSpace string) *format {
	render.ColorSpace = colorSpace
	return f
}

func (f *format) Render() *RenderFormat {
	if render.FrameRate == 0 {
		panic("frame rate must be set")
	}
	frameRate := Common.FrameDuration(render.FrameRate)
	render.Name = fmt.Sprintf("FFVideoFormat%dx%dp%.f", render.Width, render.Height, frameRate*100)
	frameDuration, err := Common.FrameMapString(render.FrameRate)
	if err != nil {
		panic(err)
	}
	render.FrameDuration = fmt.Sprintf("%ss", frameDuration)

	frameDurationCompute, _ := Common.FrameMap(render.FrameRate)
	render.FrameDurationCompute = frameDurationCompute
	return &render
}

func GetFrameRate() interface{} {
	return render.FrameRate
}

func GetFrameDuration() float64 {
	return render.FrameDurationCompute
}
