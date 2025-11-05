package image

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// CropRGBA returns a new RGBA image cropped to the specified rectangle r.
// The crop region is clamped to the source image bounds.
func CropRGBA(src *image.RGBA, r image.Rectangle) *image.RGBA {
	r = r.Intersect(src.Bounds())
	dst := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	xdraw.Draw(dst, dst.Bounds(), src, r.Min, xdraw.Src)
	return dst
}

// ResizeRGBA scales an image to the specified width (W) and height (H)
// using Catmull-Rom resampling and returns the result as an RGBA image.
func ResizeRGBA(src image.Image, W, H int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, W, H))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

// LoadImage opens and decodes a PNG or JPEG image efficiently.
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %q: %w", path, err)
	}
	return img, nil
}

// ToRGBA converts an image.Image to *image.RGBA efficiently.
// If the source image is already *image.RGBA, it returns it directly.
func ToRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}

	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}

// ExportPNG writes img as PNG with given compression level.
func ExportPNG(img image.Image, path string, level png.CompressionLevel) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	return (&png.Encoder{CompressionLevel: level}).Encode(f, img)
}

// ExportJPEG writes img as JPEG with given quality (1–100).
func ExportJPEG(img image.Image, path string, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// ExportAuto chooses encoder by file extension (.png, .jpg, .jpeg).
func ExportAuto(img image.Image, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return ExportPNG(img, path, png.BestSpeed)
	case ".jpg", ".jpeg":
		return ExportJPEG(img, path, 75)
	default:
		return fmt.Errorf("unsupported extension %q", ext)
	}
}
