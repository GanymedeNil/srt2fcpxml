// Package extract reads the titles and captions on a Final Cut Pro timeline
// from an FCPXML document, so they can be written out as subtitles.
//
// It reads every FCPXML 1.x version. Timing follows Apple's FCPXML timing rules:
// a child's offset is in its parent's local time, an anchored item's offset is
// in its base element's local time (which begins at the base's start), and a
// spine shares its parent's timeline. Contained items are cut to their parent's
// visible range; anchored items only to their parent's container.
package extract

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Kind tells titles from captions.
type Kind string

const (
	Title   Kind = "title"
	Caption Kind = "caption"
)

// Cue is one title or caption, timed relative to the start of the project.
type Cue struct {
	Start, End time.Duration
	Text       string
	Kind       Kind
}

// Options selects what to extract.
type Options struct {
	// Project names the project to read; it is required when the document holds more than one.
	Project string
	// Only keeps titles or captions; "" keeps both.
	Only Kind
}

// Result is what Extract found.
type Result struct {
	Project  string
	Cues     []Cue
	Warnings []string
}

// ReadPath extracts from a .fcpxml file or a .fcpxmld bundle (a directory with Info.fcpxml inside).
func ReadPath(path string, opts Options) (Result, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return Result{}, err
	}
	if fi.IsDir() {
		path = filepath.Join(path, "Info.fcpxml")
	}
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	return Extract(f, opts)
}

// Extract reads an FCPXML document.
func Extract(r io.Reader, opts Options) (Result, error) {
	if opts.Only != "" && opts.Only != Title && opts.Only != Caption {
		return Result{}, fmt.Errorf("unknown kind %q (use title or caption)", opts.Only)
	}
	root, err := parse(r)
	if err != nil {
		return Result{}, err
	}
	if root.name != "fcpxml" {
		return Result{}, fmt.Errorf("not an FCPXML document (root element <%s>)", root.name)
	}

	project, err := pickProject(root, opts.Project)
	if err != nil {
		return Result{}, err
	}
	seq := project.child("sequence")
	if seq == nil {
		return Result{}, fmt.Errorf("project %q has no sequence", project.attr("name"))
	}

	w := walker{media: map[string]*node{}, only: opts.Only, warned: map[string]bool{}}
	if res := root.child("resources"); res != nil {
		for _, m := range res.childrenNamed("media") {
			w.media[m.attr("id")] = m
		}
	}
	tcStart, err := w.time(seq, "tcStart")
	if err != nil {
		return Result{}, err
	}
	window := span{start: tcStart}
	if d, err := w.time(seq, "duration"); err == nil && d.Sign() > 0 {
		window.end = new(big.Rat).Add(tcStart, d)
	}
	identity := mapping{base: new(big.Rat), start: new(big.Rat)}
	if spine := seq.child("spine"); spine != nil {
		w.walkChildren(spine, identity, window, window)
	}
	if w.err != nil {
		return Result{}, w.err
	}

	cues := make([]Cue, 0, len(w.cues))
	for _, c := range w.cues {
		start := toDuration(new(big.Rat).Sub(c.start, tcStart))
		end := toDuration(new(big.Rat).Sub(c.end, tcStart))
		if end > start {
			cues = append(cues, Cue{Start: start, End: end, Text: c.text, Kind: c.kind})
		}
	}
	sort.SliceStable(cues, func(i, j int) bool { return cues[i].Start < cues[j].Start })
	return Result{Project: project.attr("name"), Cues: cues, Warnings: w.warnings}, nil
}

func pickProject(root *node, name string) (*node, error) {
	var projects []*node
	var collect func(n *node)
	collect = func(n *node) {
		for _, c := range n.children {
			switch c.name {
			case "project":
				projects = append(projects, c)
			case "library", "event":
				collect(c)
			}
		}
	}
	collect(root)
	if len(projects) == 0 {
		return nil, errors.New("the document has no project")
	}
	var names []string
	for _, p := range projects {
		if name != "" && p.attr("name") == name {
			return p, nil
		}
		names = append(names, fmt.Sprintf("%q", p.attr("name")))
	}
	if name != "" {
		return nil, fmt.Errorf("no project named %q (projects: %s)", name, strings.Join(names, ", "))
	}
	if len(projects) > 1 {
		return nil, fmt.Errorf("the document has %d projects, choose one: %s", len(projects), strings.Join(names, ", "))
	}
	return projects[0], nil
}

// mapping converts an element's local time t to timeline time: base + (t - start).
type mapping struct {
	base, start *big.Rat
}

func (m mapping) at(t *big.Rat) *big.Rat {
	r := new(big.Rat).Sub(t, m.start)
	return r.Add(r, m.base)
}

// span is a range of timeline time; a nil end means it is open-ended.
type span struct {
	start, end *big.Rat
}

func (s span) intersect(o span) span {
	r := s
	if o.start != nil && (r.start == nil || o.start.Cmp(r.start) > 0) {
		r.start = o.start
	}
	if o.end != nil && (r.end == nil || o.end.Cmp(r.end) < 0) {
		r.end = o.end
	}
	return r
}

type rawCue struct {
	start, end *big.Rat
	text       string
	kind       Kind
}

type walker struct {
	media    map[string]*node
	only     Kind
	cues     []rawCue
	warnings []string
	warned   map[string]bool
	inMedia  []string // media being walked, to stop reference cycles
	err      error
}

// Story elements whose anchored or contained items are walked.
var storyElements = map[string]bool{
	"gap": true, "clip": true, "asset-clip": true, "sync-clip": true, "ref-clip": true,
	"video": true, "audio": true, "title": true, "caption": true, "audition": true,
	"live-drawing": true, "spine": true,
}

// walkChildren walks the story elements inside parent. m maps parent's local
// time to the timeline; contained is the window for contained items and
// anchored the window for anchored ones.
func (w *walker) walkChildren(parent *node, m mapping, contained, anchored span) {
	for _, c := range parent.children {
		if w.err != nil {
			return
		}
		if !storyElements[c.name] {
			continue
		}
		window := contained
		if lane := c.attr("lane"); lane != "" && lane != "0" {
			window = anchored
		}
		w.walk(c, m, window)
	}
}

// walk handles one story element. m maps its parent's local time to the
// timeline, and window is the part of the timeline it can be seen in.
func (w *walker) walk(n *node, parent mapping, window span) {
	if n.attr("enabled") == "0" {
		return
	}
	if n.name == "spine" {
		// A spine has no start: its children share the parent's timeline.
		w.walkChildren(n, parent, window, window)
		return
	}
	if n.name == "audition" {
		// The first child is the active pick.
		for _, c := range n.children {
			if storyElements[c.name] {
				w.walk(c, parent, window)
				break
			}
		}
		return
	}

	offset, err1 := w.time(n, "offset")
	start, err2 := w.time(n, "start")
	if err := errors.Join(err1, err2); err != nil {
		w.err = err
		return
	}
	m := mapping{base: parent.at(offset), start: start}
	visible := span{start: m.base}
	if d, err := w.time(n, "duration"); err == nil && d.Sign() > 0 {
		visible.end = new(big.Rat).Add(m.base, d)
	}
	visible = visible.intersect(window)

	if n.child("timeMap") != nil || n.child("conform-rate") != nil {
		w.warn(fmt.Sprintf("%s %q is retimed; times of titles connected to it may be off", n.name, n.attr("name")))
	}

	switch n.name {
	case "title", "caption":
		kind := Kind(n.name)
		if (w.only == "" || w.only == kind) && visible.end != nil && visible.end.Cmp(visible.start) > 0 {
			if text := textOf(n); text != "" {
				w.cues = append(w.cues, rawCue{start: visible.start, end: visible.end, text: text, kind: kind})
			}
		}
	case "ref-clip":
		w.walkMedia(n.attr("ref"), m, visible)
	}
	w.walkChildren(n, m, visible, window)
}

// walkMedia walks the sequence of a compound clip's media.
func (w *walker) walkMedia(id string, m mapping, visible span) {
	for _, open := range w.inMedia {
		if open == id {
			return
		}
	}
	media := w.media[id]
	if media == nil {
		return
	}
	seq := media.child("sequence")
	if seq == nil {
		return // multicam
	}
	spine := seq.child("spine")
	if spine == nil {
		return
	}
	w.inMedia = append(w.inMedia, id)
	w.walkChildren(spine, m, visible, visible)
	w.inMedia = w.inMedia[:len(w.inMedia)-1]
}

func (w *walker) warn(msg string) {
	if !w.warned[msg] {
		w.warned[msg] = true
		w.warnings = append(w.warnings, msg)
	}
}

// time reads a time attribute such as "1001/30000s" or "5s"; a missing one is 0.
func (w *walker) time(n *node, attr string) (*big.Rat, error) {
	v := n.attr(attr)
	if v == "" {
		return new(big.Rat), nil
	}
	r, ok := new(big.Rat).SetString(strings.TrimSuffix(v, "s"))
	if !ok {
		return nil, fmt.Errorf("<%s %s=%q>: not a time value", n.name, attr, v)
	}
	return r, nil
}

// toDuration rounds rational seconds to whole milliseconds.
func toDuration(r *big.Rat) time.Duration {
	ms := new(big.Rat).Mul(r, big.NewRat(1000, 1))
	n := new(big.Int).Quo(ms.Num(), ms.Denom())
	rem := new(big.Rat).Sub(ms, new(big.Rat).SetInt(n))
	if rem.Cmp(big.NewRat(1, 2)) >= 0 {
		n.Add(n, big.NewInt(1))
	} else if rem.Cmp(big.NewRat(-1, 2)) <= 0 {
		n.Sub(n, big.NewInt(1))
	}
	return time.Duration(n.Int64()) * time.Millisecond
}

// textOf joins the text blocks of a title or caption, one per line.
func textOf(n *node) string {
	var blocks []string
	for _, t := range n.childrenNamed("text") {
		var b strings.Builder
		for _, p := range t.parts {
			switch v := p.(type) {
			case string:
				// Text directly in <text>; whitespace-only runs are indentation.
				if strings.TrimSpace(v) != "" {
					b.WriteString(v)
				}
			case *node:
				b.WriteString(v.allText())
			}
		}
		if s := strings.TrimSpace(normalizeNewlines(b.String())); s != "" {
			blocks = append(blocks, s)
		}
	}
	return strings.Join(blocks, "\n")
}

func normalizeNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// node is an XML element that keeps its text and children in document order.
type node struct {
	name     string
	attrs    map[string]string
	children []*node
	parts    []any // string or *node
}

func (n *node) attr(name string) string { return n.attrs[name] }

func (n *node) child(name string) *node {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}
	return nil
}

func (n *node) childrenNamed(name string) []*node {
	var out []*node
	for _, c := range n.children {
		if c.name == name {
			out = append(out, c)
		}
	}
	return out
}

func (n *node) allText() string {
	var b strings.Builder
	for _, p := range n.parts {
		switch v := p.(type) {
		case string:
			b.WriteString(v)
		case *node:
			b.WriteString(v.allText())
		}
	}
	return b.String()
}

func parse(r io.Reader) (*node, error) {
	dec := xml.NewDecoder(r)
	var root *node
	var stack []*node
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading FCPXML: %w", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			n := &node{name: t.Name.Local, attrs: map[string]string{}}
			for _, a := range t.Attr {
				n.attrs[a.Name.Local] = a.Value
			}
			if len(stack) == 0 {
				if root != nil {
					return nil, errors.New("reading FCPXML: more than one root element")
				}
				root = n
			} else {
				p := stack[len(stack)-1]
				p.children = append(p.children, n)
				p.parts = append(p.parts, n)
			}
			stack = append(stack, n)
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				p := stack[len(stack)-1]
				p.parts = append(p.parts, string(t))
			}
		}
	}
	if root == nil {
		return nil, errors.New("reading FCPXML: empty document")
	}
	return root, nil
}
