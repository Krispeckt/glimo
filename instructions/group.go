// Package instructions provides primitives for grouping and drawing bounded shapes together.
package instructions

import (
	"image"
	"math"

	"github.com/Krispeckt/glimo/internal/core/geom"
	"golang.org/x/image/draw"
)

// ContainerMode defines Figma-like container behavior.
type ContainerMode int

const (
	// GroupMode : children use direct overlay coordinates. No group offset applied.
	GroupMode ContainerMode = iota
	// FrameMode : children use local coordinates of the frame and are offset by (g.x, g.y).
	FrameMode
)

// Group represents a collection of drawable shapes rendered as a composite.
type Group struct {
	x, y   int           // Frame top-left (used only in FrameMode)
	w, h   int           // Frame size; if 0, computed from content bounds (FrameMode)
	clip   bool          // Clip to frame rect in FrameMode
	mode   ContainerMode // FrameMode or GroupMode
	shapes []BoundedShape
}

// NewGroup creates a new Group with GroupMode by default.
func NewGroup() *Group { return &Group{mode: GroupMode} }

// Position returns the current frame position (relevant in FrameMode).
func (g *Group) Position() (int, int) { return g.x, g.y }

// SetPosition sets frame position (relevant in FrameMode).
func (g *Group) SetPosition(x, y int) { g.x, g.y = x, y }

// SetPositionChain sets position and returns the group for chaining.
func (g *Group) SetPositionChain(x, y int) *Group { g.x, g.y = x, y; return g }

// SetMode sets container mode.
func (g *Group) SetMode(m ContainerMode) *Group { g.mode = m; return g }

// Mode returns container mode.
func (g *Group) Mode() ContainerMode { return g.mode }

// SetFrameSize sets explicit frame size for FrameMode. Zero means auto from content.
func (g *Group) SetFrameSize(w, h int) *Group { g.w, g.h = w, h; return g }

// SetClip enables or disables clipping to the frame rect in FrameMode.
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
// - FrameMode: explicit frame size if set, otherwise content bounds size.
// - GroupMode: content bounds size.
func (g *Group) Size() *geom.Size {
	if g == nil {
		return geom.NewSize(0, 0)
	}
	if g.mode == FrameMode && g.w > 0 && g.h > 0 {
		return geom.NewSize(float64(g.w), float64(g.h))
	}
	r, ok := g.bounds()
	if !ok {
		return geom.NewSize(0, 0)
	}
	return geom.NewSize(float64(r.Dx()), float64(r.Dy()))
}

// Draw renders shapes onto overlay according to the container mode.
func (g *Group) Draw(base, overlay *image.RGBA) {
	if g == nil || overlay == nil || len(g.shapes) == 0 {
		return
	}

	// Determine target rect and offset.
	rect := overlay.Bounds()
	offX, offY := 0, 0

	if g.mode == FrameMode {
		// Frame rect: explicit or from content bounds, then offset by frame position.
		if g.w > 0 && g.h > 0 {
			rect = image.Rect(g.x, g.y, g.x+g.w, g.y+g.h)
		} else if local, ok := g.bounds(); ok {
			rect = local.Add(image.Pt(g.x, g.y))
		} else {
			rect = image.Rect(g.x, g.y, g.x, g.y) // empty
		}
		offX, offY = g.x, g.y
	} else {
		// GroupMode: direct coordinates, no offset.
		rect = overlay.Bounds()
	}

	// Composite into tmp, then blit into overlay within rect.
	tmp := image.NewRGBA(overlay.Bounds())

	for _, s := range g.shapes {
		if s == nil {
			continue
		}
		sx, sy := s.Position()
		// Apply offset only in FrameMode.
		s.SetPosition(sx+offX, sy+offY)

		layer := image.NewRGBA(overlay.Bounds())
		s.Draw(base, layer)

		// Restore original position.
		s.SetPosition(sx, sy)

		draw.Draw(tmp, overlay.Bounds(), layer, layer.Bounds().Min, draw.Over)
	}

	// Clip to frame bounds in FrameMode if requested by copying only rect.
	// In GroupMode rect == overlay.Bounds(), so this is a full copy.
	if g.mode == FrameMode && g.clip {
		draw.Draw(overlay, rect, tmp, rect.Min, draw.Over)
	} else {
		// No extra clip: draw full composite.
		draw.Draw(overlay, overlay.Bounds(), tmp, tmp.Bounds().Min, draw.Over)
	}
}
