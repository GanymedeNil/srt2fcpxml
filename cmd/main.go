package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"srt2fcpxml/core"
	"strconv"
	"strings"

	"github.com/asticode/go-astisub"
)

func main() {
	srtFile := flag.String("srt", "", "srt subtitle file")
	frameDurationPoint := flag.String("fd", "25", "frame rate currently supported 23.98、24、25、29.97、30、50、59.94、60")
	width := flag.Int("width", 1920, "width resolution default 1920")
	height := flag.Int("height", 1080, "high resolution default 1080")

	flag.Parse()
	var frameDuration interface{}
	if len(*frameDurationPoint) > 2 {
		frameDuration, _ = strconv.ParseFloat(*frameDurationPoint, 64)
	} else {
		frameDuration, _ = strconv.Atoi(*frameDurationPoint)
	}

	f, _ := astisub.OpenFile(*srtFile)
	out := `<?xml version="1.0" encoding="UTF-8" ?>
	<!DOCTYPE fcpxml>
	
	`
	if len(*srtFile) == 0 {
		flag.PrintDefaults()
		os.Exit(20)
	}

	project, path := getPath(*srtFile)
	result, _ := core.Srt2FcpXmlExport(project, frameDuration, f, *width, *height)
	out += string(result)
	targetFile := fmt.Sprintf("%s/%s.fcpxml", path, project)
	fd, err := os.Create(targetFile)
	defer fd.Close()
	if err != nil {
		fmt.Println(err)
	}
	_, err = fd.Write([]byte(out))
	if err != nil {
		fmt.Println(err)
	}
}

func getPath(filePath string) (projectName, targetPath string) {
	path, _ := filepath.Abs(filePath)
	parts := strings.Split(path, "/")
	projectName = func(file string) string {
		parts := strings.Split(file, ".")
		return strings.Join(parts[0:len(parts)-1], ".")
	}(parts[len(parts)-1])
	targetPath = func(parts []string) string {
		return strings.Join(parts, "/")
	}(parts[0 : len(parts)-1])
	return
}
