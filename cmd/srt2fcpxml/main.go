// Command srt2fcpxml converts a subtitle file (SRT, WebVTT, SSA/ASS, TTML or STL)
// into an FCPXML project for Final Cut Pro, with each subtitle as a Basic Title.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/convert"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/asticode/go-astisub"
)

// version is set at release time with -ldflags "-X main.version=...".
var version string

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// errUsage marks errors caused by how the command was called.
var errUsage = errors.New("usage")

func run(args []string, stdout, stderr io.Writer) int {
	err := convertFile(args, stdout, stderr)
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	case errors.Is(err, errUsage):
		return 2
	default:
		fmt.Fprintln(stderr, "srt2fcpxml:", err)
		return 1
	}
}

func convertFile(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("srt2fcpxml", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: srt2fcpxml [flags] <subtitle file>")
		fmt.Fprintln(stderr, "Converts SRT, WebVTT, SSA/ASS, TTML or STL subtitles to a Final Cut Pro project (fcpxml).")
		fmt.Fprintln(stderr, "\nFlags:")
		fs.PrintDefaults()
	}
	defaults := convert.DefaultStyle()
	var (
		input       = fs.String("srt", "", "subtitle file (can also be given as the argument)")
		output      = fs.String("o", "", "output file, or - for stdout (default: next to the subtitle file, with an .fcpxml extension)")
		frameRate   = fs.String("fd", "25", "frame rate: "+fcpxml.SupportedRates)
		width       = fs.Int("width", 1920, "video width in pixels")
		height      = fs.Int("height", 1080, "video height in pixels")
		styleFile   = fs.String("style", "", "JSON style file; the -font*, -stroke-* and -position-y flags override it")
		font        = fs.String("font", defaults.Font, "font family")
		fontFace    = fs.String("font-face", defaults.FontFace, "font face, e.g. Regular, Bold")
		fontSize    = fs.Float64("font-size", defaults.FontSize, "font size (Basic Title template at 1920x1080)")
		fontColor   = fs.String("font-color", "#FFFFFF", `text color: "#RRGGBB", "#RRGGBBAA" or "r g b a" (0-1)`)
		strokeColor = fs.String("stroke-color", "", "outline color (needs -stroke-width)")
		strokeWidth = fs.Float64("stroke-width", 0, "outline width")
		positionY   = fs.Float64("position-y", defaults.PositionY, "vertical position from the center of the frame; negative is lower")
		showVersion = fs.Bool("version", false, "print the version and exit")
	)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return errUsage
	}
	if *showVersion {
		fmt.Fprintln(stdout, "srt2fcpxml", versionString())
		return nil
	}

	switch {
	case *input == "" && fs.NArg() == 1:
		*input = fs.Arg(0)
	case *input == "" || fs.NArg() > 0:
		fs.Usage()
		return errUsage
	}

	rate, err := fcpxml.ParseRate(*frameRate)
	if err != nil {
		return err
	}

	style := defaults
	if *styleFile != "" {
		if style, err = convert.LoadStyle(*styleFile, style); err != nil {
			return err
		}
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "font":
			style.Font = *font
		case "font-face":
			style.FontFace = *fontFace
		case "font-size":
			style.FontSize = *fontSize
		case "font-color":
			style.FontColor = *fontColor
		case "stroke-color":
			style.StrokeColor = *strokeColor
		case "stroke-width":
			style.StrokeWidth = *strokeWidth
		case "position-y":
			style.PositionY = *positionY
		}
	})

	subtitles, err := astisub.OpenFile(*input)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", *input, err)
	}

	abs, err := filepath.Abs(*input)
	if err != nil {
		return err
	}
	base := filepath.Base(abs)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	doc, err := convert.Convert(subtitles, convert.Options{
		ProjectName:     name,
		Rate:            rate,
		Width:           *width,
		Height:          *height,
		Style:           style,
		LibraryLocation: defaultLibrary(name),
	})
	if err != nil {
		return fmt.Errorf("%s: %w", *input, err)
	}
	out, err := doc.Marshal()
	if err != nil {
		return err
	}

	switch *output {
	case "-":
		_, err = stdout.Write(out)
		return err
	case "":
		*output = filepath.Join(filepath.Dir(abs), name+".fcpxml")
	}
	return os.WriteFile(*output, out, 0o644)
}

// defaultLibrary is the library Final Cut Pro is asked to import into: ~/Movies/<name>.fcpbundle.
func defaultLibrary(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(filepath.Join(home, "Movies", name+".fcpbundle"))}
	return u.String()
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
