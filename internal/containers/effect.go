package containers

import (
	"image"

	"github.com/Krispeckt/glimo/effects"
)

// Effects manages a sequence of visual effects and applies them
// in correct order relative to the rendering pipeline.
type Effects struct {
	list []effects.Effect
}

// Add appends a single effect to the list.
func (c *Effects) Add(e effects.Effect) *Effects {
	c.list = append(c.list, e)
	return c
}

// AddList appends multiple effects to the list.
func (c *Effects) AddList(es []effects.Effect) *Effects {
	c.list = append(c.list, es...)
	return c
}

// PreApplyAll applies all pre-render effects (IsPre() == true) to dst.
func (c *Effects) PreApplyAll(dst *image.RGBA) {
	if len(c.list) == 0 {
		return
	}
	for _, e := range c.list {
		if e != nil && e.IsPre() {
			e.Apply(dst)
		}
	}
}

// PostApplyAll applies all post-render effects (IsPre() == false) to dst.
func (c *Effects) PostApplyAll(dst *image.RGBA) {
	if len(c.list) == 0 {
		return
	}
	for _, e := range c.list {
		if e != nil && !e.IsPre() {
			e.Apply(dst)
		}
	}
}

// Clear removes all registered effects.
func (c *Effects) Clear() *Effects {
	c.list = nil
	return c
}

// Count returns the number of stored effects.
func (c *Effects) Count() int { return len(c.list) }
