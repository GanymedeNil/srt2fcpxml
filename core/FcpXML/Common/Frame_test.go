package Common

import (
	"log"
	"srt2fcpxml/lib"
	"testing"
)

func TestFrameMap(t *testing.T) {
	type args struct {
		frameRate interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:    "23.98",
			args:    struct{ frameRate interface{} }{frameRate: 23.98},
			want:    1001 / float64(24000),
			wantErr: false,
		},
		{
			name:    "24",
			args:    struct{ frameRate interface{} }{frameRate: 24},
			want:    100 / float64(2400),
			wantErr: false,
		},
		{
			name:    "err",
			args:    struct{ frameRate interface{} }{frameRate: "ad"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FrameMap(tt.args.frameRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("FrameMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			f := lib.Floater{Accuracy: 0.000000000001}
			log.Print(got)
			log.Print(tt.want)
			if !f.IsEqual(got, tt.want) {
				t.Errorf("FrameMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFrameMapString(t *testing.T) {
	type args struct {
		frameRate interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "23.98",
			args: struct{ frameRate interface{} }{frameRate: 23.98},
			want: "1001/24000",
		},
		{
			name:    "24",
			args:    struct{ frameRate interface{} }{frameRate: 24},
			want:    "100/2400",
			wantErr: false,
		},
		{
			name:    "err",
			args:    struct{ frameRate interface{} }{frameRate: "ad"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FrameMapString(tt.args.frameRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("FrameMapString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FrameMapString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
