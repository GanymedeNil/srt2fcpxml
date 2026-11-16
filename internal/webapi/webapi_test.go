package webapi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func read(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func decode(t *testing.T, s string) Response {
	t.Helper()
	var r Response
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		t.Fatalf("bad response %q: %v", s, err)
	}
	return r
}

func TestConvert(t *testing.T) {
	srt := read(t, "../convert/testdata/basic.srt")
	en := read(t, "../convert/testdata/basic-en.srt")
	req := `{"name":"talk.srt","secondaryName":"talk.en.srt","fps":"23.976","width":1080,"height":1920,
		"style":{"fontSize":60,"secondary":{"fontColor":"#FFD700"}}}`
	r := decode(t, Convert(req, srt, en))
	if r.Error != "" {
		t.Fatal(r.Error)
	}
	if r.Filename != "talk.fcpxml" {
		t.Errorf("filename = %q", r.Filename)
	}
	for _, want := range []string{`<project name="talk"`, `frameDuration="1001/24000s" width="1080" height="1920"`,
		`fontSize="60"`, `fontColor="1 0.843137 0 1"`, `>Hello, world</text-style>`, "<library>"} {
		if !strings.Contains(r.Output, want) {
			t.Errorf("output lacks %s", want)
		}
	}
}

func TestConvertSplitLines(t *testing.T) {
	r := decode(t, Convert(`{"name":"b.srt","fps":"25","width":1920,"height":1080,"splitLines":true}`,
		read(t, "../convert/testdata/bilingual-split.srt"), nil))
	if r.Error != "" || !strings.Contains(r.Output, `fontSize="36"`) {
		t.Errorf("split lines: %+v", r)
	}
}

func TestConvertErrors(t *testing.T) {
	srt := read(t, "../convert/testdata/basic.srt")
	for name, req := range map[string]string{
		"not json":       `{`,
		"bad fps":        `{"name":"a.srt","fps":"120","width":1920,"height":1080}`,
		"bad style":      `{"name":"a.srt","fps":"25","width":1920,"height":1080,"style":{"fontSzie":1}}`,
		"bad format":     `{"name":"a.docx","fps":"25","width":1920,"height":1080}`,
		"bad resolution": `{"name":"a.srt","fps":"25","width":0,"height":1080}`,
	} {
		if r := decode(t, Convert(req, srt, nil)); r.Error == "" || r.Output != "" {
			t.Errorf("%s: expected an error, got %+v", name, r)
		}
	}
}

func TestExtract(t *testing.T) {
	doc := read(t, "../extract/testdata/structures.fcpxml")
	r := decode(t, Extract(`{"name":"structures.fcpxml"}`, doc))
	if r.Error != "" || r.Filename != "structures.srt" || !strings.Contains(r.Output, "Anchored on clip") || len(r.Warnings) != 1 {
		t.Errorf("srt: %+v", r)
	}
	r = decode(t, Extract(`{"name":"Info.fcpxml","format":".vtt","only":"captions"}`, doc))
	if r.Error != "" || r.Filename != "Main.vtt" || !strings.HasPrefix(r.Output, "WEBVTT") || strings.Contains(r.Output, "Anchored") {
		t.Errorf("vtt from a bundle's Info.fcpxml: %+v", r)
	}
	two := read(t, "../extract/testdata/two-projects.fcpxml")
	if r := decode(t, Extract(`{"name":"x.fcpxml"}`, two)); !strings.Contains(r.Error, `"First", "Second"`) {
		t.Errorf("two projects: %+v", r)
	}
	if r := decode(t, Extract(`{"name":"x.fcpxml","project":"Second"}`, two)); r.Error != "" || !strings.Contains(r.Output, "two") {
		t.Errorf("chosen project: %+v", r)
	}
}
