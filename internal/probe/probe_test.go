package probe

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func parseFixture(t *testing.T, name string) Info {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	info, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

// The fixtures are ffprobe 8.1 output for videos made with ffmpeg's test source.
func TestParse(t *testing.T) {
	tests := []struct {
		file string
		want Info
	}{
		{"ntsc24_tc.mov", Info{ // 23.976 fps, timecode 01:00:00:00, stereo audio longer than the video
			FrameRateNum: 24000, FrameRateDen: 1001, Width: 1920, Height: 1080,
			Duration: 2961292 * time.Microsecond, Timecode: "01:00:00:00",
			HasAudio: true, AudioRate: 48000, AudioChannels: 2,
		}},
		{"phone_rot.mp4", Info{ // stored 1920x1080 with a 90° display matrix, as phones record portrait video
			FrameRateNum: 30, FrameRateDen: 1, Width: 1080, Height: 1920,
			Duration: 2 * time.Second,
		}},
		{"df2997.mov", Info{ // 29.97 fps with drop-frame timecode
			FrameRateNum: 30000, FrameRateDen: 1001, Width: 1280, Height: 720,
			Duration: 2002 * time.Millisecond, Timecode: "00:59:59;28",
		}},
	}
	for _, tt := range tests {
		if got := parseFixture(t, tt.file); got != tt.want {
			t.Errorf("%s:\n got %+v\nwant %+v", tt.file, got, tt.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for name, data := range map[string]string{
		"not json":    `nope`,
		"audio only":  `{"streams":[{"codec_type":"audio","sample_rate":"48000","channels":2}],"format":{"duration":"1"}}`,
		"bad rate":    `{"streams":[{"codec_type":"video","r_frame_rate":"0/0","width":10,"height":10,"duration":"1"}]}`,
		"no duration": `{"streams":[{"codec_type":"video","r_frame_rate":"25/1","width":10,"height":10}]}`,
	} {
		if _, err := Parse([]byte(data)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

// Runs the real ffprobe on a generated clip when FFmpeg is installed.
func TestProbeWithFFmpeg(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not installed")
	}
	path := filepath.Join(t.TempDir(), "clip.mov")
	out, err := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i", "testsrc2=size=640x360:rate=25:duration=1",
		"-timecode", "10:00:00:00", "-pix_fmt", "yuv420p", path).CombinedOutput()
	if err != nil {
		t.Skipf("ffmpeg could not make a test clip: %v\n%s", err, out)
	}
	info, err := Probe(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.FrameRate() != "25/1" || info.Width != 640 || info.Height != 360 || info.Timecode != "10:00:00:00" || info.Duration != time.Second {
		t.Errorf("unexpected probe result: %+v", info)
	}
	if _, err := Probe(filepath.Join(t.TempDir(), "missing.mov")); err == nil {
		t.Error("probing a missing file should fail")
	}
}
