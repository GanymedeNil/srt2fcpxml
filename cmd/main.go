package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"srt2fcpxml/core"
	"srt2fcpxml/core/FcpXML/Common"
	"strings"

	"github.com/asticode/go-astisub"
)

const xmlHeader = "<?xml version=\"1.0\" encoding=\"UTF-8\" ?>\n<!DOCTYPE fcpxml>\n\n"

func main() {
	srtFile := flag.String("srt", "", "srt subtitle file")
	frameRate := flag.String("fd", "25", "frame rate currently supported "+Common.SupportedRates)
	width := flag.Int("width", 1920, "width resolution default 1920")
	height := flag.Int("height", 1080, "high resolution default 1080")

	flag.Parse()
	if len(*srtFile) == 0 {
		flag.PrintDefaults()
		os.Exit(20)
	}

	if err := run(*srtFile, *frameRate, *width, *height); err != nil {
		fmt.Fprintln(os.Stderr, "srt2fcpxml:", err)
		os.Exit(1)
	}
}

func run(srtFile, frameRate string, width, height int) error {
	rate, err := Common.ParseRate(frameRate)
	if err != nil {
		return err
	}

	subtitles, err := astisub.OpenFile(srtFile)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", srtFile, err)
	}

	project, path, err := getPath(srtFile)
	if err != nil {
		return err
	}
	result, err := core.Srt2FcpXmlExport(project, rate, subtitles, width, height)
	if err != nil {
		return fmt.Errorf("%s: %w", srtFile, err)
	}

	targetFile := filepath.Join(path, project+".fcpxml")
	if err := os.WriteFile(targetFile, append([]byte(xmlHeader), result...), 0644); err != nil {
		return err
	}
	return nil
}

func getPath(filePath string) (projectName, targetPath string, err error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", err
	}
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base)), filepath.Dir(path), nil
}
