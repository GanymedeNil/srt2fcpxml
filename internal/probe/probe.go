// Package probe reads a video's frame rate, size, duration, start timecode and
// audio layout with ffprobe (part of FFmpeg).
package probe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Info describes a video file.
type Info struct {
	// FrameRate is the stream's frame rate as num/den, e.g. 24000/1001.
	FrameRateNum, FrameRateDen int64
	// Width and Height are the displayed size, with any 90° rotation applied.
	Width, Height int
	Duration      time.Duration
	// Timecode is the timecode of the first frame ("01:00:00:00"; ';' before the
	// frames means drop-frame), or "" when the file has none.
	Timecode      string
	HasAudio      bool
	AudioRate     int
	AudioChannels int
}

// FrameRate returns the frame rate as written by ffprobe, e.g. "24000/1001".
func (i Info) FrameRate() string {
	return fmt.Sprintf("%d/%d", i.FrameRateNum, i.FrameRateDen)
}

// Probe runs ffprobe on path.
func Probe(path string) (Info, error) {
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		return Info{}, errors.New("ffprobe not found: install FFmpeg (https://ffmpeg.org), or leave out -video and set -fd, -width and -height")
	}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(ffprobe, "-v", "error", "-print_format", "json", "-show_streams", "-show_format", path)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return Info{}, fmt.Errorf("ffprobe %s: %s", path, msg)
		}
		return Info{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	return Parse(stdout.Bytes())
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType    string            `json:"codec_type"`
		CodecTag     string            `json:"codec_tag_string"`
		Width        int               `json:"width"`
		Height       int               `json:"height"`
		RFrameRate   string            `json:"r_frame_rate"`
		Duration     string            `json:"duration"`
		SampleRate   string            `json:"sample_rate"`
		Channels     int               `json:"channels"`
		Tags         map[string]string `json:"tags"`
		SideDataList []struct {
			Rotation *float64 `json:"rotation"`
		} `json:"side_data_list"`
		Disposition struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
	} `json:"streams"`
	Format struct {
		Duration string            `json:"duration"`
		Tags     map[string]string `json:"tags"`
	} `json:"format"`
}

// Parse reads the JSON that "ffprobe -print_format json -show_streams -show_format" prints.
func Parse(data []byte) (Info, error) {
	var out ffprobeOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return Info{}, fmt.Errorf("reading ffprobe output: %w", err)
	}
	var info Info
	video := -1
	for i, s := range out.Streams {
		switch s.CodecType {
		case "video":
			// Skip cover art and thumbnails stored as video streams.
			if video < 0 && s.Disposition.AttachedPic == 0 {
				video = i
			}
		case "audio":
			if !info.HasAudio {
				info.HasAudio = true
				info.AudioChannels = s.Channels
				info.AudioRate, _ = strconv.Atoi(s.SampleRate)
			}
		}
	}
	if video < 0 {
		return Info{}, errors.New("the file has no video stream")
	}
	v := out.Streams[video]

	num, den, ok := parseFraction(v.RFrameRate)
	if !ok {
		return Info{}, fmt.Errorf("unknown frame rate %q", v.RFrameRate)
	}
	info.FrameRateNum, info.FrameRateDen = num, den

	info.Width, info.Height = v.Width, v.Height
	for _, sd := range v.SideDataList {
		if sd.Rotation != nil && int(*sd.Rotation)%180 != 0 {
			info.Width, info.Height = v.Height, v.Width
		}
	}
	if r, err := strconv.Atoi(v.Tags["rotate"]); err == nil && r%180 != 0 {
		info.Width, info.Height = v.Height, v.Width
	}

	// The video stream's own duration, so a slightly longer audio track does not
	// add frames without picture.
	// ffprobe prints decimal seconds ("2.002000"); ParseDuration reads them exactly.
	d, err := time.ParseDuration(v.Duration + "s")
	if err != nil {
		if d, err = time.ParseDuration(out.Format.Duration + "s"); err != nil {
			return Info{}, errors.New("the video has no duration")
		}
	}
	info.Duration = d

	info.Timecode = v.Tags["timecode"]
	if info.Timecode == "" {
		info.Timecode = out.Format.Tags["timecode"]
	}
	if info.Timecode == "" {
		for _, s := range out.Streams {
			if s.CodecTag == "tmcd" && s.Tags["timecode"] != "" {
				info.Timecode = s.Tags["timecode"]
				break
			}
		}
	}
	return info, nil
}

func parseFraction(s string) (num, den int64, ok bool) {
	n, d, found := strings.Cut(s, "/")
	if !found {
		d = "1"
	}
	num, err1 := strconv.ParseInt(n, 10, 64)
	den, err2 := strconv.ParseInt(d, 10, 64)
	if err1 != nil || err2 != nil || num <= 0 || den <= 0 {
		return 0, 0, false
	}
	return num, den, true
}
