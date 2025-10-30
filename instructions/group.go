// Package instructions provides primitives for grouping and drawing bounded shapes together.
package instructions

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"golang.org/x/image/draw"
)

// Group represents a frame-like container of drawable shapes rendered as a composite.
// Children use local coordinates and are offset by (x, y). Optional clipping to the frame.
type Group struct {
	x, y   int  // Frame top-left
	w, h   int  // Frame size; if 0, computed from content bounds
	clip   bool // Clip to frame rect
	shapes []BoundedShape
}

// NewGroup creates a new Group with frame semantics by default.
func NewGroup() *Group { return &Group{} }

// Position returns the current frame position.
func (g *Group) Position() (int, int) { return g.x, g.y }

// SetPosition sets frame position.
func (g *Group) SetPosition(x, y int) { g.x, g.y = x, y }

// SetPositionChain sets position and returns the group for chaining.
func (g *Group) SetPositionChain(x, y int) *Group { g.x, g.y = x, y; return g }

// SetFrameSize sets explicit frame size. Zero means auto from content.
func (g *Group) SetFrameSize(w, h int) *Group { g.w, g.h = w, h; return g }

// SetClip enables or disables clipping to the frame rect.
func (g *Group) SetClip(clip bool) *Group { g.clip = clip; return g }

// AddInstruction adds a single shape.
func (g *Group) AddInstruction(s BoundedShape) {
	if g != nil && s != nil {
		g.shapes = append(g.shapes, s)
	}
}

// AddInstructions adds multiple shapes.
func (g *Group) AddInstructions(shapes ...BoundedShape) {
	if g == nil {
		return
	}
	for _, s := range shapes {
		if s != nil {
			g.shapes = append(g.shapes, s)
		}
	}
}

// Clear removes all shapes.
func (g *Group) Clear() {
	if g != nil {
		g.shapes = g.shapes[:0]
	}
}

// bounds computes union of child bounds using each shape's Position() and Size().
// Coordinates are local to the frame (no offset by g.x, g.y).
func (g *Group) bounds() (r image.Rectangle, ok bool) {
	if g == nil || len(g.shapes) == 0 {
		return
	}
	first := true
	for _, s := range g.shapes {
		if s == nil || s.Size() == nil {
			continue
		}
		sx, sy := s.Position()
		w := int(math.Ceil(math.Max(0, s.Size().Width())))
		h := int(math.Ceil(math.Max(0, s.Size().Height())))
		if w == 0 || h == 0 {
			continue
		}
		sr := image.Rect(sx, sy, sx+w, sy+h)
		if first {
			r, first = sr, false
		} else {
			r = r.Union(sr)
		}
	}
	ok = !first && r.Dx() > 0 && r.Dy() > 0
	return
}

// Size returns composite size.
// - Explicit frame size if set, otherwise content bounds size.
func (g *Group) Size() *geom.Size {
	if g == nil {
		return geom.NewSize(0, 0)
	}
	if g.w > 0 && g.h > 0 {
		return geom.NewSize(float64(g.w), float64(g.h))
	}
	r, ok := g.bounds()
	if !ok {
		return geom.NewSize(0, 0)
	}
	return geom.NewSize(float64(r.Dx()), float64(r.Dy()))
}

// cloneBaseTo allocates an RGBA with given bounds and copies overlapping pixels from src.
func cloneBaseTo(bounds image.Rectangle, src *image.RGBA) *image.RGBA {
	acc := image.NewRGBA(bounds)
	if src == nil {
		return acc
	}
	copyRect := bounds.Intersect(src.Bounds())
	if !copyRect.Empty() {
		draw.Draw(acc, copyRect, src, copyRect.Min, draw.Src)
	}
	return acc
}

func (g *Group) Draw(base, overlay *image.RGBA) {
	if g == nil || overlay == nil || len(g.shapes) == 0 {
		return
	}

	// Frame rect.
	var frameRect image.Rectangle
	if g.w > 0 && g.h > 0 {
		frameRect = image.Rect(g.x, g.y, g.x+g.w, g.y+g.h)
	} else if local, ok := g.bounds(); ok {
		frameRect = local.Add(image.Pt(g.x, g.y))
	} else {
		return
	}

	dst := overlay.Bounds()

	// Work window: full dst when no clip, otherwise frame∩dst.
	work := dst
	if g.clip {
		work = frameRect.Intersect(dst)
		if work.Empty() {
			return
		}
	}

	target := image.NewRGBA(work)
	acc := cloneBaseTo(work, base)

	draw.Draw(target, target.Bounds(), acc, acc.Bounds().Min, draw.Src)

	offX, offY := g.x, g.y
	var dirty image.Rectangle

	for _, s := range g.shapes {
		if s == nil || s.Size() == nil {
			continue
		}
		sx, sy := s.Position()
		sw := int(math.Ceil(math.Max(0, s.Size().Width())))
		sh := int(math.Ceil(math.Max(0, s.Size().Height())))
		if sw == 0 || sh == 0 {
			continue
		}
		abs := image.Rect(sx+offX, sy+offY, sx+offX+sw, sy+offY+sh)
		changed := abs.Intersect(work)
		if changed.Empty() {
			continue
		}

		s.SetPosition(sx+offX, sy+offY)
		s.Draw(acc, target)
		s.SetPosition(sx, sy)

		draw.Draw(acc, changed, target, changed.Min, draw.Src)

		if dirty.Empty() {
			dirty = changed
		} else {
			dirty = dirty.Union(changed)
		}
	}

	if !dirty.Empty() {
		draw.Draw(overlay, dirty, target, dirty.Min, draw.Src)
	}
}
