package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtures = "../../internal/extract/testdata/"

func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

// copyFixture copies a fixture into a temp dir, so the default output lands there.
func copyFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(fixtures + name)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWritesSRTNextToInput(t *testing.T) {
	in := copyFixture(t, "structures.fcpxml")
	code, _, stderr := runCLI(in)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if !strings.Contains(stderr, "warning: clip \"Slow\" is retimed") {
		t.Errorf("retiming warning missing: %q", stderr)
	}
	out, err := os.ReadFile(strings.TrimSuffix(in, ".fcpxml") + ".srt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "00:00:02,000 --> 00:00:04,000\nAnchored on clip") {
		t.Errorf("unexpected SRT:\n%s", out)
	}
}

func TestOutputFormatsAndFlags(t *testing.T) {
	code, stdout, stderr := runCLI("-o", "-", "-only", "captions", fixtures+"structures.fcpxml")
	if code != 0 || !strings.Contains(stdout, "A caption") || strings.Contains(stdout, "Anchored on clip") {
		t.Errorf("-only captions: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}

	vtt := filepath.Join(t.TempDir(), "out.vtt")
	if code, _, stderr := runCLI("-o", vtt, "-project", "Second", fixtures+"two-projects.fcpxml"); code != 0 {
		t.Fatalf("-o .vtt: exit %d: %s", code, stderr)
	}
	if data, _ := os.ReadFile(vtt); !strings.HasPrefix(string(data), "WEBVTT") || !strings.Contains(string(data), "two") {
		t.Errorf("unexpected VTT:\n%s", data)
	}

	code, stdout, _ = runCLI("-o", "-", fixtures+"bundle.fcpxmld")
	if code != 0 || !strings.Contains(stdout, "From a bundle") {
		t.Errorf("bundle: exit %d, stdout %q", code, stdout)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		args     []string
		code     int
		inStderr string
	}{
		{nil, 2, "Usage"},
		{[]string{"-nope", "x"}, 2, "flag provided but not defined"},
		{[]string{"missing.fcpxml"}, 1, "no such file"},
		{[]string{fixtures + "two-projects.fcpxml"}, 1, `choose one: "First", "Second"`},
		{[]string{"-o", "x.ass", fixtures + "bundle.fcpxmld"}, 1, "unsupported output format"},
		{[]string{"-only", "markers", fixtures + "bundle.fcpxmld"}, 1, "unknown kind"},
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
	if code != 0 || !strings.HasPrefix(stdout, "fcpxml2srt ") {
		t.Errorf("exit %d, stdout %q", code, stdout)
	}
}
