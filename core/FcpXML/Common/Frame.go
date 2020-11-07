package Common

import (
	"errors"
	"fmt"
	"srt2fcpxml/lib"
)

func FrameMapString(frameRate interface{}) (string, error) {
	switch v := frameRate.(type) {
	case float64:
		return fmt.Sprintf("%d/%d", 1001, int64(lib.Round(v, 0)*1000)), nil
	case int:
		return fmt.Sprintf("%d/%d", 100, v*100), nil
	default:
		return "", errors.New("frameRate is int or float")
	}
}

func FrameMap(frameRate interface{}) (float64, error) {
	switch v := frameRate.(type) {
	case float64:
		return 1001 / (lib.Round(v, 0) * 1000), nil
	case int:
		return 100 / float64(v*100), nil
	default:
		return 0, errors.New("frameRate is int or float")
	}
}

func FrameDuration(frameRate interface{}) float64 {
	switch v := frameRate.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0
	}
}

func FrameDurationFormat(frameRate interface{}) (float64, float64, error) {
	switch v := frameRate.(type) {
	case float64:
		return 1001, lib.Round(v, 0) * 1000, nil
	case int:
		return 100, float64(v * 100), nil
	default:
		return 0, 0, errors.New("frameRate is int or float")
	}
}
