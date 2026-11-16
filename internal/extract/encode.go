package extract

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/asticode/go-astisub"
)

// Encode renders cues as subtitles; format is ".srt" or ".vtt".
func Encode(cues []Cue, format string) ([]byte, error) {
	subs := astisub.NewSubtitles()
	for _, c := range cues {
		item := &astisub.Item{StartAt: c.Start, EndAt: c.End}
		for _, line := range strings.Split(c.Text, "\n") {
			item.Lines = append(item.Lines, astisub.Line{Items: []astisub.LineItem{{Text: line}}})
		}
		subs.Items = append(subs.Items, item)
	}
	var buf bytes.Buffer
	var err error
	switch strings.ToLower(format) {
	case ".srt":
		err = subs.WriteToSRT(&buf)
	case ".vtt":
		err = subs.WriteToWebVTT(&buf)
	default:
		return nil, fmt.Errorf("unsupported output format %q (use .srt or .vtt)", format)
	}
	return buf.Bytes(), err
}
