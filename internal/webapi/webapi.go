// Package webapi is the conversion API of the web version. Requests and
// responses are JSON so the browser code stays simple; file contents are passed
// as raw bytes, so the browser and the CLI read files exactly the same way.
package webapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/convert"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/extract"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/input"
	"github.com/asticode/go-astisub"
)

// ConvertRequest asks for subtitles to be turned into an FCPXML project.
type ConvertRequest struct {
	Name          string `json:"name"`
	SecondaryName string `json:"secondaryName"`
	FrameRate     string `json:"fps"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	SplitLines    bool   `json:"splitLines"`
	// Style holds style fields to change, in the style file format.
	Style json.RawMessage `json:"style"`
}

// ExtractRequest asks for the titles and captions of an FCPXML project.
type ExtractRequest struct {
	Name    string `json:"name"`
	Project string `json:"project"`
	Only    string `json:"only"`
	// Format is ".srt" (default) or ".vtt".
	Format string `json:"format"`
}

// Response carries the converted file or an error.
type Response struct {
	Output   string   `json:"output,omitempty"`
	Filename string   `json:"filename,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Error    string   `json:"error,omitempty"`
}

// Convert handles a ConvertRequest given as JSON; secondary may be nil.
func Convert(requestJSON string, data, secondary []byte) string {
	var req ConvertRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return reply(Response{Error: "bad request: " + err.Error()})
	}
	res, err := convertFile(req, data, secondary)
	if err != nil {
		res.Error = err.Error()
	}
	return reply(res)
}

func convertFile(req ConvertRequest, data, secondary []byte) (Response, error) {
	rate, err := fcpxml.ParseRate(req.FrameRate)
	if err != nil {
		return Response{}, err
	}
	style := convert.DefaultStyle()
	if len(req.Style) > 0 && string(req.Style) != "null" {
		if style, err = convert.ParseStyle(req.Style, style); err != nil {
			return Response{}, fmt.Errorf("style: %w", err)
		}
	}
	subs, err := input.Read(req.Name, bytes.NewReader(data))
	if err != nil {
		return Response{}, fmt.Errorf("%s: %w", req.Name, err)
	}
	var secondarySubs *astisub.Subtitles
	if secondary != nil {
		if secondarySubs, err = input.Read(req.SecondaryName, bytes.NewReader(secondary)); err != nil {
			return Response{}, fmt.Errorf("%s: %w", req.SecondaryName, err)
		}
	}
	name := strings.TrimSuffix(req.Name, filepath.Ext(req.Name))
	var res Response
	doc, err := convert.Convert(subs, convert.Options{
		ProjectName: name,
		Rate:        rate,
		Width:       req.Width,
		Height:      req.Height,
		Style:       style,
		Secondary:   secondarySubs,
		SplitLines:  req.SplitLines,
		// The browser cannot know the user's home folder, so the library is left
		// without a location.
		Warnf: func(format string, args ...any) { res.Warnings = append(res.Warnings, fmt.Sprintf(format, args...)) },
	})
	if err != nil {
		return res, err
	}
	out, err := doc.Marshal()
	if err != nil {
		return res, err
	}
	res.Output, res.Filename = string(out), name+".fcpxml"
	return res, nil
}

// Extract handles an ExtractRequest given as JSON.
func Extract(requestJSON string, data []byte) string {
	var req ExtractRequest
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return reply(Response{Error: "bad request: " + err.Error()})
	}
	format := strings.ToLower(req.Format)
	if format == "" {
		format = ".srt"
	}
	result, err := extract.Extract(bytes.NewReader(data), extract.Options{
		Project: req.Project,
		Only:    extract.Kind(strings.TrimSuffix(req.Only, "s")),
	})
	if err != nil {
		return reply(Response{Error: err.Error()})
	}
	res := Response{Warnings: result.Warnings}
	if len(result.Cues) == 0 {
		res.Error = fmt.Sprintf("no titles or captions found in project %q", result.Project)
		return reply(res)
	}
	out, err := extract.Encode(result.Cues, format)
	if err != nil {
		res.Error = err.Error()
		return reply(res)
	}
	name := strings.TrimSuffix(req.Name, filepath.Ext(req.Name))
	if name == "" || name == "Info" { // Info.fcpxml from inside a .fcpxmld bundle
		name = result.Project
	}
	res.Output, res.Filename = string(out), name+format
	return reply(res)
}

func reply(r Response) string {
	b, _ := json.Marshal(r)
	return string(b)
}
