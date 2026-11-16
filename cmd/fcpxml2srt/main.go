// Command fcpxml2srt exports the titles and captions on a Final Cut Pro
// timeline (an .fcpxml file or .fcpxmld bundle) as SRT or WebVTT subtitles.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/extract"
)

// version is set at release time with -ldflags "-X main.version=...".
var version string

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// errUsage marks errors caused by how the command was called.
var errUsage = errors.New("usage")

func run(args []string, stdout, stderr io.Writer) int {
	err := exportFile(args, stdout, stderr)
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	case errors.Is(err, errUsage):
		return 2
	default:
		fmt.Fprintln(stderr, "fcpxml2srt:", err)
		return 1
	}
}

func exportFile(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("fcpxml2srt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: fcpxml2srt [flags] <project.fcpxml | project.fcpxmld>")
		fmt.Fprintln(stderr, "Exports the titles and captions on a Final Cut Pro timeline as SRT or WebVTT subtitles.")
		fmt.Fprintln(stderr, "\nFlags:")
		fs.PrintDefaults()
	}
	var (
		output      = fs.String("o", "", "output file: .srt or .vtt, or - for SRT on stdout (default: next to the input, with an .srt extension)")
		project     = fs.String("project", "", "project to export, when the file holds more than one")
		only        = fs.String("only", "", "export only titles or only captions: title, caption")
		showVersion = fs.Bool("version", false, "print the version and exit")
	)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return errUsage
	}
	if *showVersion {
		fmt.Fprintln(stdout, "fcpxml2srt", versionString())
		return nil
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return errUsage
	}
	input := fs.Arg(0)

	res, err := extract.ReadPath(input, extract.Options{Project: *project, Only: extract.Kind(strings.TrimSuffix(*only, "s"))})
	if err != nil {
		return fmt.Errorf("%s: %w", input, err)
	}
	for _, w := range res.Warnings {
		fmt.Fprintln(stderr, "fcpxml2srt: warning:", w)
	}
	if len(res.Cues) == 0 {
		return fmt.Errorf("%s: no titles or captions found in project %q", input, res.Project)
	}

	target := *output
	if target == "" {
		abs, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		base := filepath.Base(abs)
		target = filepath.Join(filepath.Dir(abs), strings.TrimSuffix(base, filepath.Ext(base))+".srt")
	}
	format := strings.ToLower(filepath.Ext(target))
	if target == "-" {
		format = ".srt"
	}
	out, err := extract.Encode(res.Cues, format)
	if err != nil {
		return err
	}
	if target == "-" {
		_, err = stdout.Write(out)
		return err
	}
	return os.WriteFile(target, out, 0o644)
}

func versionString() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}
