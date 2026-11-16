package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const srt = "1\n00:00:01,000 --> 00:00:02,000\nHello\n"

// setup writes a subtitle file into a temp dir and returns its path.
func setup(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestWritesNextToInput(t *testing.T) {
	in := setup(t, "my talk.srt", srt)
	for _, args := range [][]string{{"-srt", in}, {in}} {
		code, _, stderr := runCLI(args...)
		if code != 0 {
			t.Fatalf("%v: exit %d: %s", args, code, stderr)
		}
		out, err := os.ReadFile(filepath.Join(filepath.Dir(in), "my talk.fcpxml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), `<project name="my talk"`) {
			t.Errorf("project not named after the file:\n%s", out)
		}
	}
}

func TestStdoutAndOutputFlag(t *testing.T) {
	in := setup(t, "a.srt", srt)
	code, stdout, stderr := runCLI("-o", "-", in)
	if code != 0 || !strings.HasPrefix(stdout, "<?xml") {
		t.Fatalf("-o -: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	target := filepath.Join(t.TempDir(), "out.fcpxml")
	if code, _, stderr := runCLI("-o", target, in); code != 0 {
		t.Fatalf("-o file: exit %d: %s", code, stderr)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
}

func TestStyleFlagsOverrideStyleFile(t *testing.T) {
	in := setup(t, "a.srt", srt)
	style := setup(t, "style.json", `{"font": "Helvetica", "fontSize": 60}`)
	code, stdout, stderr := runCLI("-o", "-", "-style", style, "-font-size", "70", "-stroke-color", "#000000", "-stroke-width", "2", in)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, want := range []string{`font="Helvetica"`, `fontSize="70"`, `strokeColor="0 0 0 1"`, `strokeWidth="2"`} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output lacks %s", want)
		}
	}
}

func TestErrors(t *testing.T) {
	in := setup(t, "a.srt", srt)
	empty := setup(t, "empty.srt", "")
	badStyle := setup(t, "style.json", `{"fontSzie": 60}`)
	tests := []struct {
		args     []string
		code     int
		inStderr string
	}{
		{nil, 2, "Usage"},
		{[]string{in, in}, 2, "Usage"},
		{[]string{"-nope", in}, 2, "flag provided but not defined"},
		{[]string{"-fd", "120", in}, 1, "unsupported frame rate"},
		{[]string{"missing.srt"}, 1, "cannot read"},
		{[]string{empty}, 1, "no subtitles found"},
		{[]string{"-width", "0", in}, 1, "invalid resolution"},
		{[]string{"-style", badStyle, in}, 1, "unknown field"},
		{[]string{"-stroke-color", "#000", in}, 1, "strokeColor"},
	}
	for _, tt := range tests {
		code, _, stderr := runCLI(tt.args...)
		if code != tt.code || !strings.Contains(stderr, tt.inStderr) {
			t.Errorf("%v: exit %d, stderr %q; want exit %d with %q", tt.args, code, stderr, tt.code, tt.inStderr)
		}
	}
}

func TestVersion(t *testing.T) {
	code, stdout, _ := runCLI("-version")
	if code != 0 || !strings.HasPrefix(stdout, "srt2fcpxml ") {
		t.Errorf("exit %d, stdout %q", code, stdout)
	}
}

func TestBilingual(t *testing.T) {
	zh := setup(t, "zh.srt", "1\n00:00:01,000 --> 00:00:02,000\n你好\n")
	en := setup(t, "en.srt", "1\n00:00:01,000 --> 00:00:02,000\n<i>Hello</i> <b>there</b>\n")
	code, stdout, stderr := runCLI("-o", "-", "-secondary", en, "-secondary-font-size", "30", "-secondary-font-color", "#FFD700", zh)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	for _, want := range []string{`>你好&#xA;</text-style>`, `>Hello</text-style>`, `> </text-style>`, `>there</text-style>`, `fontSize="30"`, `fontColor="1 0.843137 0 1"`} {
		if !strings.Contains(stdout, want) {
			t.Errorf("output lacks %s", want)
		}
	}

	both := setup(t, "both.srt", "1\n00:00:01,000 --> 00:00:02,000\n你好\nHello\n")
	code, stdout, stderr = runCLI("-o", "-", "-split-lines", both)
	if code != 0 || !strings.Contains(stdout, `>你好&#xA;</text-style>`) || !strings.Contains(stdout, `fontSize="36"`) {
		t.Errorf("-split-lines: exit %d, stderr %q", code, stderr)
	}
}

func TestBilingualErrors(t *testing.T) {
	in := setup(t, "a.srt", srt)
	if code, _, stderr := runCLI("-secondary", in, "-split-lines", in); code != 1 || !strings.Contains(stderr, "not both") {
		t.Errorf("both options: exit %d, stderr %q", code, stderr)
	}
	if code, _, stderr := runCLI("-secondary", "missing.srt", in); code != 1 || !strings.Contains(stderr, "cannot read missing.srt") {
		t.Errorf("missing secondary: exit %d, stderr %q", code, stderr)
	}
}
