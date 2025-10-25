package glimo

import (
	"image"

	"github.com/Krispeckt/glimo/instructions"
	imageUtil "github.com/Krispeckt/glimo/internal/core/image"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/Krispeckt/glimo/internal/render"
)

//
// Package Overview
//
// The `glimo` package provides a unified public API for image and graphics rendering.
// It consolidates internal modules under a concise namespace for drawing, font management,
// and color handling.
//

//
// Type Aliases
//
// These aliases re-export key internal types to simplify the public interface.
//

type (
	// Font represents a loaded font resource for text rendering.
	Font = render.Font
	// Color defines the RGBA color model used throughout rendering and fill operations.
	Color = patterns.Color
	// Layer represents a 2D drawable surface.
	Layer = instructions.Layer
	// Frame is an alias for Layer, used semantically for frame-based rendering.
	Frame = instructions.Layer
)

//
// Layer Constructors
//
// Layers are RGBA canvases that can be drawn on, blended, or exported.
// Supported image formats for loading are PNG and JPEG.
//

// NewLayer creates a blank RGBA layer with the specified width and height.
func NewLayer(width, height int) *instructions.Layer {
	return instructions.NewLayer(width, height)
}

// NewLayerFromImage wraps an existing image.Image into a Layer.
func NewLayerFromImage(img image.Image) *instructions.Layer {
	return instructions.NewLayerFromImage(img)
}

// NewLayerFromRGBA constructs a Layer directly from an existing RGBA buffer.
func NewLayerFromRGBA(rgba *image.RGBA) *instructions.Layer {
	return instructions.NewLayerFromRGBA(rgba)
}

// NewLayerFromImagePath loads a PNG or JPEG image from a given file path and returns it as a Layer.
func NewLayerFromImagePath(path string) (*instructions.Layer, error) {
	return instructions.NewLayerFromImagePath(path)
}

// MustLoadLayerFromImagePath loads a Layer from an image file and panics on failure.
func MustLoadLayerFromImagePath(path string) *instructions.Layer {
	return instructions.MustLoadLayerFromImagePath(path)
}

//
// Frame Constructors
//
// Frames are functionally identical to Layers but are used for animation or sequential rendering.
//

// NewFrame creates a blank frame surface with given dimensions.
func NewFrame(width, height int) *instructions.Layer {
	return instructions.NewLayer(width, height)
}

// NewFrameFromImage wraps an existing image.Image as a frame.
func NewFrameFromImage(img image.Image) *instructions.Layer {
	return instructions.NewLayerFromImage(img)
}

// NewFrameFromRGBA creates a frame from an RGBA buffer.
func NewFrameFromRGBA(rgba *image.RGBA) *instructions.Layer {
	return instructions.NewLayerFromRGBA(rgba)
}

// NewFrameFromImagePath loads a PNG or JPEG image and returns it as a frame.
func NewFrameFromImagePath(path string) (*instructions.Layer, error) {
	return instructions.NewLayerFromImagePath(path)
}

// MustLoadFrameFromImagePath loads a frame from a file and panics on failure.
func MustLoadFrameFromImagePath(path string) *instructions.Layer {
	return instructions.MustLoadLayerFromImagePath(path)
}

//
// Font Management
//
// Functions for loading and managing fonts through the internal render subsystem.
//

// LoadFont loads a font from a file path.
func LoadFont(path string, sizePt float64) (*render.Font, error) {
	return render.LoadFont(path, sizePt)
}

// LoadFontFromBytes loads a font from a byte slice in memory.
func LoadFontFromBytes(data []byte, sizePt float64) (*render.Font, error) {
	return render.LoadFontFromBytes(data, sizePt)
}

// MustLoadFont loads a font and panics if loading fails.
func MustLoadFont(path string, sizePt float64) *render.Font {
	return render.MustLoadFont(path, sizePt)
}

// MustLoadFontFromBytes loads a font from memory and panics on failure.
func MustLoadFontFromBytes(data []byte, sizePt float64) *render.Font {
	return render.MustLoadFontFromBytes(data, sizePt)
}

// SetFontCacheCapacity sets the maximum number of cached faces for memory management.
func SetFontCacheCapacity(limit int) {
	render.SetFontCacheCapacity(limit)
}

// ClearFontCache removes all cached font data.
func ClearFontCache() {
	render.ClearFontCache()
}

//
// Image Utilities
//
// General-purpose helpers for image loading, decoding, and format conversion.
//

// LoadImage loads an image from a file and returns it as image.Image.
func LoadImage(path string) (image.Image, error) {
	return imageUtil.LoadImage(path)
}
