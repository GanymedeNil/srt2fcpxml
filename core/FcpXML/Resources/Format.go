package Resources

import (
	"fmt"
	"srt2fcpxml/core/FcpXML/Common"
)

type format struct {
	frameRate  Common.Rate
	width      int
	height     int
	colorSpace string
}

type RenderFormat struct {
	Text          string `xml:",chardata"`
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	FrameDuration string `xml:"frameDuration,attr"`
	Width         int    `xml:"width,attr"`
	Height        int    `xml:"height,attr"`
	ColorSpace    string `xml:"colorSpace,attr"`
}

func NewFormat() *format {
	return &format{
		width:      1920,
		height:     1080,
		colorSpace: "1-1-1 (Rec. 709)",
	}
}

func (f *format) SetWidth(width int) *format {
	f.width = width
	return f
}
func (f *format) SetHeight(height int) *format {
	f.height = height
	return f
}
func (f *format) SetFrameRate(frameRate Common.Rate) *format {
	f.frameRate = frameRate
	return f
}

func (f *format) SetColorSpace(colorSpace string) *format {
	f.colorSpace = colorSpace
	return f
}

func (f *format) Render() *RenderFormat {
	if f.frameRate.IsZero() {
		panic("frame rate must be set")
	}
	return &RenderFormat{
		ID:            "r1",
		Name:          fmt.Sprintf("FFVideoFormat%dx%dp%.f", f.width, f.height, f.frameRate.FPS()*100),
		FrameDuration: f.frameRate.FrameDuration(),
		Width:         f.width,
		Height:        f.height,
		ColorSpace:    f.colorSpace,
	}
}
