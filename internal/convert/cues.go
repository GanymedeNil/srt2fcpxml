package convert

import (
	"errors"
	"sort"
	"time"

	"github.com/asticode/go-astisub"
)

// A cue is one subtitle on the timeline: primary-language lines above
// secondary-language lines. Either may be empty.
type cue struct {
	start, end time.Duration
	primary    []astisub.Line
	secondary  []astisub.Line
}

// buildCues turns the subtitles into cues. With opts.SplitLines the first line
// of each subtitle is primary and the rest secondary. With opts.Secondary each
// secondary subtitle joins the primary subtitle it overlaps most in time;
// one that overlaps none becomes a cue of its own.
func buildCues(subs *astisub.Subtitles, opts Options) ([]cue, error) {
	if opts.Secondary != nil && opts.SplitLines {
		return nil, errors.New("use either a secondary subtitle file or split lines, not both")
	}
	cues := make([]cue, 0, len(subs.Items))
	for _, item := range subs.Items {
		c := cue{start: item.StartAt, end: item.EndAt, primary: item.Lines}
		if opts.SplitLines && len(item.Lines) > 1 {
			c.primary, c.secondary = item.Lines[:1], item.Lines[1:]
		}
		cues = append(cues, c)
	}
	if opts.Secondary == nil {
		return cues, nil
	}
	if len(opts.Secondary.Items) == 0 {
		return nil, errors.New("no subtitles found in the secondary subtitles")
	}

	var unmatched []cue
	for _, s := range opts.Secondary.Items {
		best, bestOverlap := -1, time.Duration(0)
		for i, c := range cues {
			if overlap := min(c.end, s.EndAt) - max(c.start, s.StartAt); overlap > bestOverlap {
				best, bestOverlap = i, overlap
			}
		}
		if best < 0 {
			unmatched = append(unmatched, cue{start: s.StartAt, end: s.EndAt, secondary: s.Lines})
			continue
		}
		cues[best].secondary = append(cues[best].secondary, s.Lines...)
	}
	if len(unmatched) > 0 {
		cues = append(cues, unmatched...)
		sort.SliceStable(cues, func(i, j int) bool { return cues[i].start < cues[j].start })
	}
	return cues, nil
}
