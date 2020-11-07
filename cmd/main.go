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
	srtFile := flag.String("srt", "", "srt 字幕文件")
	frameDurationPoint := flag.String("fd", "25", "帧率目前支持 23.98、24、25、29.97、30、50、59.94、60")
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
	result, _ := core.Srt2FcpXmlExport(project, frameDuration, f)
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
