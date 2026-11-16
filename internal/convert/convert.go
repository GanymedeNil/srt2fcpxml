// Package convert turns subtitles into an FCPXML project whose subtitles are
// Basic Title clips, so they can be styled like any other title in Final Cut Pro.
package convert

import (
	"errors"
	"fmt"
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

	// Now and NewUID default to time.Now and random UUIDs; tests set them for stable output.
	Now    time.Time
	NewUID func() string
}

const (
	formatID = "r1"
	effectID = "r2"
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

	// Titles are anchored to a gap whose local timeline starts at 01:00:00:00,
	// as in projects exported by Final Cut Pro.
	gapStart := 3600 * rate.Timebase()

	// Start and end frames are each rounded from absolute times (not start plus a
	// separately rounded duration), so subtitles that touch in the SRT share a
	// frame boundary instead of overlapping or leaving a 1-frame hole. A title
	// lasts at least one frame.
	var titles []fcpxml.Title
	var endFrames int64
	texts := titleText{style: style}
	for _, c := range cues {
		startFrame, endFrame := rate.Frames(c.start), rate.Frames(c.end)
		if endFrame <= startFrame {
			endFrame = startFrame + 1
		}
		endFrames = max(endFrames, endFrame)

		text, defs, name := texts.build(c)
		titles = append(titles, fcpxml.Title{
			Name:          name,
			Lane:          "1",
			Offset:        rate.Time(gapStart + startFrame),
			Ref:           effectID,
			Duration:      rate.Time(endFrame - startFrame),
			Start:         rate.Time(gapStart),
			Params:        style.params,
			Text:          text,
			TextStyleDefs: defs,
		})
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
						Duration:    rate.Time(endFrames),
						Format:      formatID,
						TCStart:     "0s",
						TCFormat:    "NDF",
						AudioLayout: "stereo",
						AudioRate:   "48k",
						Spine: fcpxml.Spine{Gaps: []fcpxml.Gap{{
							Name:     "Gap",
							Offset:   "0s",
							Duration: rate.Time(endFrames),
							Start:    rate.Time(gapStart),
							Titles:   titles,
						}}},
					},
				}},
			}},
		},
	}, nil
}
