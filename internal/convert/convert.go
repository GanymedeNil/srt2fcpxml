// Package convert turns subtitles into an FCPXML project whose subtitles are
// Basic Title clips, so they can be styled like any other title in Final Cut Pro.
package convert

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/asticode/go-astisub"
	"github.com/google/uuid"
)

// Options controls the generated project.
type Options struct {
	// ProjectName names the project, and the library when LibraryLocation is not set.
	ProjectName string
	Rate        fcpxml.Rate
	Width       int
	Height      int
	Style       Style
	// Secondary holds second-language subtitles, shown under the primary subtitle
	// they overlap most in time, in Style.Secondary.
	Secondary *astisub.Subtitles
	// SplitLines treats the first line of each subtitle as the primary language
	// and the other lines as the secondary language.
	SplitLines bool
	// LibraryLocation is the file:// URL of the library to import into; empty leaves it out.
	LibraryLocation string
	// Video, when set, puts this clip in the primary storyline and connects the
	// subtitles to it, instead of using an empty gap.
	Video *Video
	// Warnf reports problems that do not stop the conversion; nil ignores them.
	Warnf func(format string, args ...any)

	// Now and NewUID default to time.Now and random UUIDs; tests set them for stable output.
	Now    time.Time
	NewUID func() string
}

// Video is a media file to lay the subtitles over.
type Video struct {
	Name string
	// URL is the file:// URL of the media file.
	URL      string
	Duration time.Duration
	// Timecode is the timecode of the first frame, e.g. "01:00:00:00" or
	// "00:59:59;28" (drop-frame); "" means the media starts at 0.
	Timecode      string
	HasAudio      bool
	AudioRate     int
	AudioChannels int
}

const (
	formatID = "r1"
	effectID = "r2"
	assetID  = "r3"
)

// Convert builds the FCPXML document for subtitles.
func Convert(subtitles *astisub.Subtitles, opts Options) (*fcpxml.Document, error) {
	rate := opts.Rate
	if rate.IsZero() {
		return nil, errors.New("frame rate must be set")
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		return nil, fmt.Errorf("invalid resolution %dx%d", opts.Width, opts.Height)
	}
	if subtitles == nil || len(subtitles.Items) == 0 {
		return nil, errors.New("no subtitles found")
	}
	style, err := opts.Style.compile()
	if err != nil {
		return nil, err
	}
	cues, err := buildCues(subtitles, opts)
	if err != nil {
		return nil, err
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	newUID := opts.NewUID
	if newUID == nil {
		newUID = uuid.NewString
	}

	warnf := opts.Warnf
	if warnf == nil {
		warnf = func(string, ...any) {}
	}

	// Titles are anchored to a gap whose local timeline starts at 01:00:00:00, as
	// in projects exported by Final Cut Pro, or to the video clip, whose local
	// timeline starts at the media's own timecode. A title's offset is in that
	// local time.
	gapStart := 3600 * rate.Timebase()
	anchorStart := gapStart
	var clipFrames int64
	tcFormat := "NDF"
	if v := opts.Video; v != nil {
		anchorStart = 0
		if v.Timecode != "" {
			var dropFrame bool
			if anchorStart, dropFrame, err = rate.TimecodeFrames(v.Timecode); err != nil {
				return nil, fmt.Errorf("video %s: %w", v.Name, err)
			}
			if dropFrame {
				tcFormat = "DF"
			}
		}
		if clipFrames = rate.Frames(v.Duration); clipFrames <= 0 {
			return nil, fmt.Errorf("video %s is shorter than one frame", v.Name)
		}
	}

	// Start and end frames are each rounded from absolute times (not start plus a
	// separately rounded duration), so subtitles that touch in the SRT share a
	// frame boundary instead of overlapping or leaving a 1-frame hole. A title
	// lasts at least one frame.
	var titles []fcpxml.Title
	var endFrames int64
	var pastEnd int
	texts := titleText{style: style}
	for _, c := range cues {
		startFrame, endFrame := rate.Frames(c.start), rate.Frames(c.end)
		if endFrame <= startFrame {
			endFrame = startFrame + 1
		}
		endFrames = max(endFrames, endFrame)
		if opts.Video != nil && endFrame > clipFrames {
			pastEnd++
		}

		text, defs, name := texts.build(c)
		titles = append(titles, fcpxml.Title{
			Name:          name,
			Lane:          "1",
			Offset:        rate.Time(anchorStart + startFrame),
			Ref:           effectID,
			Duration:      rate.Time(endFrame - startFrame),
			Start:         rate.Time(gapStart),
			Params:        style.params,
			Text:          text,
			TextStyleDefs: defs,
		})
	}

	sequenceFrames := endFrames
	spine := fcpxml.Spine{Gaps: []fcpxml.Gap{{
		Name:     "Gap",
		Offset:   "0s",
		Duration: rate.Time(endFrames),
		Start:    rate.Time(gapStart),
		Titles:   titles,
	}}}
	var assets []fcpxml.Asset
	if v := opts.Video; v != nil {
		if pastEnd > 0 {
			warnf("%d subtitle(s) end after the video ends (%s)", pastEnd, v.Duration.Round(time.Millisecond))
		}
		sequenceFrames = clipFrames
		asset := fcpxml.Asset{
			ID:       assetID,
			Name:     v.Name,
			Src:      v.URL,
			Start:    rate.Time(anchorStart),
			Duration: rate.Time(clipFrames),
			HasVideo: "1",
			Format:   formatID,
		}
		if v.HasAudio {
			asset.HasAudio, asset.AudioSources = "1", "1"
			if v.AudioChannels > 0 {
				asset.AudioChannels = strconv.Itoa(v.AudioChannels)
			}
			if v.AudioRate > 0 {
				asset.AudioRate = strconv.Itoa(v.AudioRate)
			}
		}
		assets = append(assets, asset)
		spine = fcpxml.Spine{AssetClips: []fcpxml.AssetClip{{
			Ref:      assetID,
			Offset:   "0s",
			Name:     v.Name,
			Start:    rate.Time(anchorStart),
			Duration: rate.Time(clipFrames),
			Format:   formatID,
			TCFormat: tcFormat,
			Titles:   titles,
		}}}
	}

	return &fcpxml.Document{
		Version: fcpxml.Version,
		Resources: fcpxml.Resources{
			Formats: []fcpxml.Format{{
				ID:            formatID,
				Name:          fcpxml.PredefinedFormatName(rate, opts.Width, opts.Height),
				FrameDuration: rate.FrameDuration(),
				Width:         opts.Width,
				Height:        opts.Height,
				ColorSpace:    "1-1-1 (Rec. 709)",
			}},
			Effects: []fcpxml.Effect{{
				ID:   effectID,
				Name: "Basic Title",
				UID:  ".../Titles.localized/Bumper:Opener.localized/Basic Title.localized/Basic Title.moti",
			}},
			Assets: assets,
		},
		Library: fcpxml.Library{
			Location: opts.LibraryLocation,
			Events: []fcpxml.Event{{
				Name: now.Format("2006-01-02"),
				UID:  newUID(),
				Projects: []fcpxml.Project{{
					Name:    opts.ProjectName,
					UID:     newUID(),
					ModDate: now.Format("2006-01-02 15:04:05 -0700"),
					Sequence: fcpxml.Sequence{
						Duration:    rate.Time(sequenceFrames),
						Format:      formatID,
						TCStart:     "0s",
						TCFormat:    tcFormat,
						AudioLayout: "stereo",
						AudioRate:   "48k",
						Spine:       spine,
					},
				}},
			}},
		},
	}, nil
}
