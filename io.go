package prism

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
)

// decodeConfig holds decode options.
type decodeConfig struct {
	autoOrientation bool
	maxImageSize    int // max total pixels (width * height), 0 = unlimited
}

// encodeConfig holds encode options.
type encodeConfig struct {
	jpegQuality         int
	pngCompressionLevel png.CompressionLevel
	gifNumColors        int
	gifQuantizer        draw.Quantizer
	gifDrawer           draw.Drawer
}

func defaultEncodeConfig() encodeConfig {
	return encodeConfig{
		jpegQuality:         95,
		pngCompressionLevel: png.DefaultCompression,
		gifNumColors:        256,
	}
}

// AutoOrientation returns a DecodeOption that enables or disables
// automatic image rotation based on EXIF orientation data.
func AutoOrientation(enabled bool) DecodeOption {
	return func(c *decodeConfig) {
		c.autoOrientation = enabled
	}
}

// MaxImageSize returns a DecodeOption that rejects images whose total
// pixel count (width * height) exceeds the given limit.
func MaxImageSize(pixels int) DecodeOption {
	return func(c *decodeConfig) {
		c.maxImageSize = pixels
	}
}

// JPEGQuality returns an EncodeOption that sets JPEG quality (1-100).
func JPEGQuality(quality int) EncodeOption {
	return func(c *encodeConfig) {
		c.jpegQuality = quality
	}
}

// PNGCompressionLevel returns an EncodeOption that sets PNG compression level.
func PNGCompressionLevel(level png.CompressionLevel) EncodeOption {
	return func(c *encodeConfig) {
		c.pngCompressionLevel = level
	}
}

// GIFNumColors returns an EncodeOption that sets the number of colors in GIF.
func GIFNumColors(numColors int) EncodeOption {
	return func(c *encodeConfig) {
		c.gifNumColors = numColors
	}
}

// GIFQuantizer returns an EncodeOption that sets the GIF quantizer.
func GIFQuantizer(quantizer draw.Quantizer) EncodeOption {
	return func(c *encodeConfig) {
		c.gifQuantizer = quantizer
	}
}

// GIFDrawer returns an EncodeOption that sets the GIF drawer.
func GIFDrawer(drawer draw.Drawer) EncodeOption {
	return func(c *encodeConfig) {
		c.gifDrawer = drawer
	}
}

// FormatFromExtension returns the image format based on the file extension.
func FormatFromExtension(ext string) (Format, error) {
	ext = strings.ToLower(strings.TrimPrefix(ext, "."))
	switch ext {
	case "jpg", "jpeg":
		return JPEG, nil
	case "png":
		return PNG, nil
	case "gif":
		return GIF, nil
	case "tif", "tiff":
		return TIFF, nil
	case "bmp":
		return BMP, nil
	case "webp":
		return WEBP, nil
	default:
		return 0, ErrUnsupportedFormat
	}
}

// FormatFromFilename returns the image format based on the filename extension.
func FormatFromFilename(filename string) (Format, error) {
	return FormatFromExtension(filepath.Ext(filename))
}

// Open opens an image file, decodes it, and returns the image.
func Open(filename string, opts ...DecodeOption) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f, opts...)
}

// Decode reads an image from r, decodes it, and returns the image.
func Decode(r io.Reader, opts ...DecodeOption) (image.Image, error) {
	var cfg decodeConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// Enforce maximum image size.
	if cfg.maxImageSize > 0 {
		bounds := img.Bounds()
		pixels := bounds.Dx() * bounds.Dy()
		if pixels > cfg.maxImageSize {
			return nil, ErrUnsupportedFormat // image too large
		}
	}

	return img, nil
}

// Save encodes an image and saves it to a file.
// The format is determined from the filename extension.
func Save(img image.Image, filename string, opts ...EncodeOption) error {
	format, err := FormatFromFilename(filename)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return Encode(f, img, format, opts...)
}

// Encode writes an image in the specified format to w.
func Encode(w io.Writer, img image.Image, format Format, opts ...EncodeOption) error {
	cfg := defaultEncodeConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	switch format {
	case JPEG:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: cfg.jpegQuality})
	case PNG:
		enc := &png.Encoder{CompressionLevel: cfg.pngCompressionLevel}
		return enc.Encode(w, img)
	case GIF:
		return encodeGIF(w, img, cfg)
	case BMP:
		return encodeBMP(w, img)
	case TIFF:
		return encodeTIFF(w, img)
	default:
		return ErrUnsupportedFormat
	}
}

func encodeGIF(w io.Writer, img image.Image, cfg encodeConfig) error {
	bounds := img.Bounds()

	// Build palette.
	numColors := cfg.gifNumColors
	if numColors <= 0 || numColors > 256 {
		numColors = 256
	}

	var palette color.Palette
	if cfg.gifQuantizer != nil {
		palette = cfg.gifQuantizer.Quantize(make(color.Palette, 0, numColors), img)
	} else {
		// Web-safe palette as fallback.
		palette = make(color.Palette, 0, numColors)
		step := 256 / 6
		for r := 0; r < 6 && len(palette) < numColors; r++ {
			for g := 0; g < 6 && len(palette) < numColors; g++ {
				for b := 0; b < 6 && len(palette) < numColors; b++ {
					palette = append(palette, color.NRGBA{
						R: uint8(r * step), G: uint8(g * step), B: uint8(b * step), A: 255,
					})
				}
			}
		}
	}

	paletted := image.NewPaletted(bounds, palette)

	var drawer draw.Drawer
	if cfg.gifDrawer != nil {
		drawer = cfg.gifDrawer
	} else {
		drawer = draw.FloydSteinberg
	}
	drawer.Draw(paletted, bounds, img, bounds.Min)

	return gif.Encode(w, paletted, nil)
}

// encodeBMP encodes an image in BMP format.
func encodeBMP(w io.Writer, img image.Image) error {
	// BMP encoding is simple: use the x/image/bmp encoder if available,
	// otherwise fall back to manual encoding.
	// For now, return unsupported — BMP encode support added in Phase 4.
	return ErrUnsupportedFormat
}

// encodeTIFF encodes an image in TIFF format.
func encodeTIFF(w io.Writer, img image.Image) error {
	// TIFF encoding via x/image/tiff.
	// For now, return unsupported — TIFF encode support added in Phase 4.
	return ErrUnsupportedFormat
}
