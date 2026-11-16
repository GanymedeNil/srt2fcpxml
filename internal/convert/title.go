package convert

import (
	"fmt"
	"strings"

	"github.com/GanymedeNil/srt2fcpxml/v3/internal/fcpxml"
	"github.com/asticode/go-astisub"
)

// titleText builds the text of each title. Text-style-def IDs must be unique in
// the whole document, so the builder numbers them across titles.
type titleText struct {
	style  compiledStyle
	nextID int
}

// build returns the cue's text runs, the style definitions they reference and
// the title's name. Runs that look the same are merged, and each distinct look
// gets one text-style-def; lines are separated by newlines.
func (b *titleText) build(c cue) (fcpxml.Text, []fcpxml.TextStyleDef, string) {
	type segment struct {
		style fcpxml.TextStyle
		text  string
	}
	var segs []segment
	var names []string
	add := func(lines []astisub.Line, base fcpxml.TextStyle) {
		for _, line := range lines {
			if len(segs) > 0 {
				segs[len(segs)-1].text += "\n"
			}
			names = append(names, line.String())
			if len(line.Items) == 0 {
				segs = append(segs, segment{style: base})
			}
			for _, item := range line.Items {
				st := inlineStyle(item.InlineStyle).apply(base)
				if n := len(segs); n > 0 && segs[n-1].style == st {
					segs[n-1].text += item.Text
					continue
				}
				segs = append(segs, segment{style: st, text: item.Text})
			}
		}
	}
	add(c.primary, b.style.primary)
	add(c.secondary, b.style.secondary)
	if len(segs) == 0 {
		segs = append(segs, segment{style: b.style.primary})
	}

	var text fcpxml.Text
	var defs []fcpxml.TextStyleDef
	ids := map[fcpxml.TextStyle]string{}
	for _, s := range segs {
		id, ok := ids[s.style]
		if !ok {
			b.nextID++
			id = fmt.Sprintf("ts%d", b.nextID)
			ids[s.style] = id
			defs = append(defs, fcpxml.TextStyleDef{ID: id, Style: s.style})
		}
		text.Runs = append(text.Runs, fcpxml.TextRun{Ref: id, Text: s.text})
	}
	return text, defs, strings.Join(names, " - ")
}
