package instructions

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/Krispeckt/glimo/internal/core/geom"
	imageUtil "github.com/Krispeckt/glimo/internal/core/image"
	"golang.org/x/image/draw"
)

// Layer represents a 2D drawable surface backed by an RGBA buffer.
// It provides methods to draw, export, and composite images.
type Layer struct {
	x, y  int
	image *image.RGBA
	size  *geom.Size
}

// NewLayer creates a new empty Layer with the specified width and height.
// The underlying image buffer is initialized as a blank RGBA canvas.
func NewLayer(width, height int) *Layer {
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))
	return NewLayerFromRGBA(rgba)
}

// NewLayerFromImage creates a new Layer from any image.Image instance.
// The input image is converted to RGBA format if necessary.
func NewLayerFromImage(src image.Image) *Layer {
	return NewLayerFromRGBA(imageUtil.ToRGBA(src))
}

// NewLayerFromRGBA creates a new Layer from an existing *image.RGBA buffer.
// The Layer will share the same underlying pixel data as the provided buffer.
func NewLayerFromRGBA(rgba *image.RGBA) *Layer {
	return &Layer{
		image: rgba,
		size:  geom.NewSizeFromImage(rgba),
	}
}

// NewLayerFromImagePath loads an image from the given file path.
// Only PNG and JPEG formats are allowed.
func NewLayerFromImagePath(path string) (*Layer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	if format != "png" && format != "jpeg" {
		return nil, errors.New("unsupported image format: only PNG and JPEG are allowed")
	}

	return NewLayerFromRGBA(imageUtil.ToRGBA(img)), nil
}

// MustLoadLayerFromImagePath loads a PNG or JPEG image from the specified file path
// and returns it as a Layer. It panics if the file cannot be opened, decoded,
// or if the format is unsupported. Intended for initialization code,
// tests, or tools where image load failures are considered fatal.
func MustLoadLayerFromImagePath(path string) *Layer {
	l, err := NewLayerFromImagePath(path)
	if err != nil {
		panic(fmt.Errorf("MustLoadLayerFromImagePath: %w", err))
	}
	return l
}

// Size returns the dimensions of the current Layer as a *geom.Size object.
func (l *Layer) Size() *geom.Size {
	return l.size
}

// Image returns the underlying *image.RGBA buffer of the Layer.
func (l *Layer) Image() *image.RGBA {
	return l.image
}

// Position returns the top-left corner of the Layer in parent coordinates.
func (l *Layer) Position() (int, int) {
	return l.x, l.y
}

// SetPosition sets the Layer’s position in integer coordinates.
func (l *Layer) SetPosition(x, y int) {
	l.x, l.y = x, y
}

// SetPositionChain sets position and returns the same Layer for method chaining.
func (l *Layer) SetPositionChain(x, y int) *Layer {
	l.x, l.y = x, y
	return l
}

// SetSize adjusts the Layer's visible bounds to the requested width and height
// without resampling or reallocating pixels. The bounds are clamped to the
// current backing buffer capacity to avoid expansion. Content is not modified.
func (l *Layer) SetSize(w, h int) {
	if l == nil || l.image == nil {
		return
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	// Compute maximum representable width/height for the current Pix/Stride.
	maxW := 0
	if l.image.Stride > 0 {
		maxW = l.image.Stride / 4 // 4 bytes per pixel in RGBA
	}
	maxH := 0
	if l.image.Stride > 0 {
		maxH = len(l.image.Pix) / l.image.Stride
	}

	if w > maxW {
		w = maxW
	}
	if h > maxH {
		h = maxH
	}

	min := l.image.Rect.Min
	l.image.Rect = image.Rectangle{
		Min: min,
		Max: image.Pt(min.X+w, min.Y+h),
	}
	l.size = geom.NewSize(float64(w), float64(h))
}

// SetSizeChain sets the visible bounds to width and height and returns the Layer.
// Identical to SetSize but chainable.
func (l *Layer) SetSizeChain(w, h int) *Layer {
	l.SetSize(w, h)
	return l
}

// SetBounds sets the Layer’s position and visible bounds. The size change
// is clamped to the current buffer capacity and does not resample or
// reallocate pixels. Only the bounds are adjusted; content stays intact.
func (l *Layer) SetBounds(x, y, width, height int) {
	l.x, l.y = x, y
	l.SetSize(width, height)
}

// SetBoundsChain sets position and size, clamps size to buffer limits, and returns the Layer.
// Identical to SetBounds but chainable.
func (l *Layer) SetBoundsChain(x, y, width, height int) *Layer {
	l.SetBounds(x, y, width, height)
	return l
}

// Draw renders this Layer’s contents onto another base and overlay image.
//
// The method composites the layer’s RGBA data into the overlay image at
// the current (x, y) position. This allows a Layer to act as a drawable
// Shape within higher-level containers or layouts.
func (l *Layer) Draw(_, overlay *image.RGBA) {
	if l == nil || l.image == nil {
		return
	}

	r := l.image.Bounds().Add(image.Pt(l.x, l.y))
	draw.Draw(overlay, r, l.image, l.image.Bounds().Min, draw.Over)
}

// ExportPNG saves the Layer as a PNG image to the specified file path.
// The compression level controls the output file size and encoding speed.
func (l *Layer) ExportPNG(path string, level png.CompressionLevel) error {
	return imageUtil.ExportPNG(l.image, path, level)
}

// ExportJPEG saves the Layer as a JPEG image to the specified file path.
// The quality value must be between 0 and 100.
func (l *Layer) ExportJPEG(path string, quality int) error {
	if quality < 0 || quality > 100 {
		return errors.New("invalid quality level")
	}
	return imageUtil.ExportJPEG(l.image, path, quality)
}

// Export automatically determines the file format (PNG or JPEG) based on the file extension.
// Supported extensions: .png, .jpg, .jpeg.
func (l *Layer) Export(path string) error {
	return imageUtil.ExportAuto(l.image, path)
}

// ExportBytes encodes the current Layer as PNG and returns the raw byte slice.
// Useful for in-memory exports, HTTP responses, or further processing.
func (l *Layer) ExportBytes(level png.CompressionLevel) ([]byte, error) {
	var buf bytes.Buffer
	encoder := png.Encoder{CompressionLevel: level}
	if err := encoder.Encode(&buf, l.image); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LoadInstruction executes a single drawing instruction on the Layer.
// The instruction defines its own drawing behavior through the Shape interface.
func (l *Layer) LoadInstruction(shape Shape) {
	overlay := image.NewRGBA(l.image.Bounds())
	shape.Draw(l.image, overlay)

	draw.Draw(l.image, overlay.Bounds(), overlay, image.Point{}, draw.Over)
}

// LoadInstructions executes a sequence of drawing instructions in order.
// Optimized for batch operations while maintaining predictable execution order.
func (l *Layer) LoadInstructions(shapes []Shape) {
	if len(shapes) == 0 {
		return
	}
	for i := 0; i < len(shapes); i++ {
		l.LoadInstruction(shapes[i])
	}
}

// AddLayer composites another Layer on top of the current one at the specified coordinates.
// The source Layer is drawn using the Over operator, preserving transparency.
func (l *Layer) AddLayer(layer *Layer, x, y int) *Layer {
	if l == nil || l.image == nil || layer == nil || layer.image == nil {
		return l
	}
	src := layer.image
	dst := l.image
	r := src.Bounds().Add(image.Pt(x, y))
	draw.Draw(dst, r, src, src.Bounds().Min, draw.Over)
	return l
}
