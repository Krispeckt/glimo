package glimo

import (
	"github.com/Krispeckt/glimo/instructions"
	imageUtil "github.com/Krispeckt/glimo/internal/core/image"
	"github.com/Krispeckt/glimo/internal/core/image/patterns"
	"github.com/Krispeckt/glimo/internal/render"
)

// Type aliases for public API.
//
// These aliases re-export types from internal modules to present a unified
// and concise public interface under the `glimo` namespace.
type (
	Font  = render.Font        // Font resource for text rendering
	Color = patterns.Color     // Color model used for shapes and fills
	Layer = instructions.Layer // 2D drawable canvas (alias for backward compatibility)
	Frame = instructions.Layer // Alias for Layer to represent frame surfaces
)

// Constructors for creating new layers from scratch or from images.
//
// A Layer is a 2D RGBA canvas that can be drawn on, composed, or exported.
// Functions that load from paths only support PNG and JPEG formats.
var (
	// NewLayer creates a blank RGBA layer with given width and height.
	NewLayer = instructions.NewLayer

	// NewLayerFromImage creates a Layer from any image.Image.
	NewLayerFromImage = instructions.NewLayerFromImage

	// NewLayerFromRGBA creates a Layer directly from an existing RGBA buffer.
	NewLayerFromRGBA = instructions.NewLayerFromRGBA

	// NewLayerFromImagePath loads a PNG or JPEG file as a new Layer.
	NewLayerFromImagePath = instructions.NewLayerFromImagePath

	// MustLoadLayerFromImagePath loads a Layer from a file and panics on failure.
	MustLoadLayerFromImagePath = instructions.MustLoadLayerFromImagePath
)

// Frame constructors (identical to Layer, but named semantically for animation
// or frame-based rendering contexts).
var (
	NewFrame                   = instructions.NewLayer
	NewFrameFromImage          = instructions.NewLayerFromImage
	NewFrameFromRGBA           = instructions.NewLayerFromRGBA
	NewFrameFromImagePath      = instructions.NewLayerFromImagePath
	MustLoadFrameFromImagePath = instructions.MustLoadLayerFromImagePath
)

// Font management utilities.
//
// These functions provide font loading, caching, and lifecycle control
// through the internal render subsystem.
var (
	// LoadFont loads a font from a file path.
	LoadFont = render.LoadFont

	// LoadFontFromBytes loads a font directly from an in-memory byte slice.
	LoadFontFromBytes = render.LoadFontFromBytes

	// MustLoadFont loads a font and panics on failure.
	MustLoadFont = render.MustLoadFont

	// MustLoadFontFromBytes loads a font from memory and panics on failure.
	MustLoadFontFromBytes = render.MustLoadFontFromBytes

	// SetFontCacheCapacity limits the number of cached font glyphs to conserve memory.
	SetFontCacheCapacity = render.SetFontCacheCapacity

	// ClearFontCache clears all cached font data.
	ClearFontCache = render.ClearFontCache
)

// Image utility functions.
//
// These functions provide convenient access to image I/O helpers
// (loading, decoding, format conversion, etc.).
var (
	// LoadImage loads an image from a file and returns it as image.Image.
	LoadImage = imageUtil.LoadImage
)
