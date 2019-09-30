package main

import (
	"flag"
	"fmt"
	"github.com/asticode/go-astisub"
	"os"
	"srt2fcpxml/core"
	"strings"
)

func main() {
	srtFile := flag.String("srt", "", "srt 字幕文件")
	frameDuration := flag.Int("fd", 25, "帧率目前只支持整数24、25、30、50、60")
	flag.Parse()
	f, _ := astisub.OpenFile(*srtFile)
	out := `<?xml version="1.0" encoding="UTF-8" ?>
	<!DOCTYPE fcpxml>
	
	`
	if len(*srtFile)==0 {
		flag.PrintDefaults()
		os.Exit(20)
	}

	project, path := getPath(*srtFile)
	result, _ := core.Srt2FcpxmlExport(project, *frameDuration, f)
	out += string(result)
	targetFile := fmt.Sprintf("%s/%s.fcpxml", path, project)
	fd,err := os.Create(targetFile)
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
	parts := strings.Split(filePath, "/")
	projectName = func(file string) string {
		parts := strings.Split(file, ".")
		return strings.Join(parts[0:len(parts)-1], ".")
	}(parts[len(parts)-1])
	targetPath = func(parts []string) string {
		return strings.Join(parts, "/")
	}(parts[0 : len(parts)-1])
	return
}
