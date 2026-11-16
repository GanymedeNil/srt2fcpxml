// Package input reads subtitle files in every format srt2fcpxml accepts.
// The CLI and the web version both use it, so they parse files the same way.
package input

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asticode/go-astisub"
)

// Extensions lists the supported subtitle file extensions.
const Extensions = ".srt, .vtt, .ass, .ssa, .ttml, .stl, .ts"

// Open reads the subtitle file at path.
func Open(path string) (*astisub.Subtitles, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Read(path, bytes.NewReader(data))
}

// Read parses subtitles from r in the format given by name's extension.
func Read(name string, r io.Reader) (*astisub.Subtitles, error) {
	switch ext := strings.ToLower(filepath.Ext(name)); ext {
	case ".srt":
		return readSRT(r)
	case ".vtt":
		return astisub.ReadFromWebVTT(r)
	case ".ass", ".ssa":
		return astisub.ReadFromSSA(r)
	case ".ttml":
		return astisub.ReadFromTTML(r)
	case ".stl":
		return astisub.ReadFromSTL(r, astisub.STLOptions{})
	case ".ts":
		return astisub.ReadFromTeletext(r, astisub.TeletextOptions{})
	default:
		return nil, fmt.Errorf("unsupported subtitle format %q (supported: %s)", ext, Extensions)
	}
}

// keptSpace stands in for whitespace between inline tags while astisub parses SRT.
const keptSpace = ''

var spaceBetweenTags = regexp.MustCompile(`>[ \t]+<`)

// readSRT parses SRT. astisub drops text that is only whitespace, so the space in
// "<i>a</i> <b>b</b>" would be lost and the line would read "ab". The space is
// swapped for a placeholder before parsing and restored afterwards.
func readSRT(r io.Reader) (*astisub.Subtitles, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data = spaceBetweenTags.ReplaceAllFunc(data, func(m []byte) []byte {
		inner := strings.Repeat(string(keptSpace), len(m)-2)
		return []byte(">" + inner + "<")
	})
	subs, err := astisub.ReadFromSRT(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for _, item := range subs.Items {
		for _, line := range item.Lines {
			for i := range line.Items {
				line.Items[i].Text = strings.ReplaceAll(line.Items[i].Text, string(keptSpace), " ")
			}
		}
	}
	return subs, nil
}
