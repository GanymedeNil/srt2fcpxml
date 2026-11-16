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
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/input"
	"github.com/GanymedeNil/srt2fcpxml/v3/internal/probe"
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
		inputFile   = fs.String("srt", "", "subtitle file (can also be given as the argument)")
		output      = fs.String("o", "", "output file, or - for stdout (default: next to the subtitle file, with an .fcpxml extension)")
		video       = fs.String("video", "", "video file to put under the subtitles; reads -fd, -width and -height from it (needs ffprobe)")
		frameRate   = fs.String("fd", "25", "frame rate: "+fcpxml.SupportedRates)
		width       = fs.Int("width", 1920, "video width in pixels")
		height      = fs.Int("height", 1080, "video height in pixels")
		secondary   = fs.String("secondary", "", "second-language subtitle file, shown under the matching subtitle")
		splitLines  = fs.Bool("split-lines", false, "bilingual file: first line of each subtitle is the main language, the rest the second language")
		styleFile   = fs.String("style", "", "JSON style file; the -font*, -secondary-font*, -stroke-* and -position-y flags override it")
		font        = fs.String("font", defaults.Font, "font family")
		fontFace    = fs.String("font-face", defaults.FontFace, "font face, e.g. Regular, Bold")
		fontSize    = fs.Float64("font-size", defaults.FontSize, "font size (Basic Title template at 1920x1080)")
		fontColor   = fs.String("font-color", "#FFFFFF", `text color: "#RRGGBB", "#RRGGBBAA" or "r g b a" (0-1)`)
		strokeColor = fs.String("stroke-color", "", "outline color (needs -stroke-width)")
		strokeWidth = fs.Float64("stroke-width", 0, "outline width")
		positionY   = fs.Float64("position-y", defaults.PositionY, "vertical position from the center of the frame; negative is lower")
		secFont     = fs.String("secondary-font", defaults.Secondary.Font, "font family of the second language")
		secFontSize = fs.Float64("secondary-font-size", defaults.Secondary.FontSize, "font size of the second language")
		secColor    = fs.String("secondary-font-color", "#FFFFFF", "text color of the second language")
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
	case *inputFile == "" && fs.NArg() == 1:
		*inputFile = fs.Arg(0)
	case *inputFile == "" || fs.NArg() > 0:
		fs.Usage()
		return errUsage
	}

	set := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { set[f.Name] = true })

	var err error
	var videoInfo probe.Info
	if *video != "" {
		if videoInfo, err = probe.Probe(*video); err != nil {
			return err
		}
		if !set["width"] {
			*width = videoInfo.Width
		}
		if !set["height"] {
			*height = videoInfo.Height
		}
	}
	var rate fcpxml.Rate
	if *video != "" && !set["fd"] {
		var ok bool
		if rate, ok = fcpxml.RateFromFraction(videoInfo.FrameRateNum, videoInfo.FrameRateDen); !ok {
			return fmt.Errorf("%s: frame rate %s is not supported (supported: %s); choose one with -fd", *video, videoInfo.FrameRate(), fcpxml.SupportedRates)
		}
	} else if rate, err = fcpxml.ParseRate(*frameRate); err != nil {
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
		case "secondary-font":
			style.Secondary.Font = *secFont
		case "secondary-font-size":
			style.Secondary.FontSize = *secFontSize
		case "secondary-font-color":
			style.Secondary.FontColor = *secColor
		}
	})

	subtitles, err := input.Open(*inputFile)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", *inputFile, err)
	}
	var secondarySubs *astisub.Subtitles
	if *secondary != "" {
		if secondarySubs, err = input.Open(*secondary); err != nil {
			return fmt.Errorf("cannot read %s: %w", *secondary, err)
		}
	}

	abs, err := filepath.Abs(*inputFile)
	if err != nil {
		return err
	}
	base := filepath.Base(abs)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	var clip *convert.Video
	if *video != "" {
		videoPath, err := filepath.Abs(*video)
		if err != nil {
			return err
		}
		videoBase := filepath.Base(videoPath)
		clip = &convert.Video{
			Name:          strings.TrimSuffix(videoBase, filepath.Ext(videoBase)),
			URL:           fileURL(videoPath),
			Duration:      videoInfo.Duration,
			Timecode:      videoInfo.Timecode,
			HasAudio:      videoInfo.HasAudio,
			AudioRate:     videoInfo.AudioRate,
			AudioChannels: videoInfo.AudioChannels,
		}
	}

	doc, err := convert.Convert(subtitles, convert.Options{
		ProjectName:     name,
		Rate:            rate,
		Width:           *width,
		Height:          *height,
		Style:           style,
		Secondary:       secondarySubs,
		SplitLines:      *splitLines,
		Video:           clip,
		LibraryLocation: defaultLibrary(name),
		Warnf: func(format string, args ...any) {
			fmt.Fprintf(stderr, "srt2fcpxml: warning: "+format+"\n", args...)
		},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", *inputFile, err)
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
	return fileURL(filepath.Join(home, "Movies", name+".fcpbundle"))
}

// fileURL turns an absolute path into a file:// URL, escaping spaces and other characters.
func fileURL(path string) string {
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
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
