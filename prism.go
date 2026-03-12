// Package prism provides image processing functions.
// It is a drop-in replacement for disintegration/imaging v1.6.2.
package prism

import (
	"errors"
	"image"
	"image/color"
	"math"
)

// Format is an image file format.
type Format int

const (
	JPEG Format = iota
	PNG
	GIF
	TIFF
	BMP
	WEBP
)

// Anchor is a point in a 2D rectangle used for alignment operations.
type Anchor int

const (
	Center Anchor = iota
	TopLeft
	Top
	TopRight
	Left
	Right
	BottomLeft
	Bottom
	BottomRight
)

// ResampleFilter defines an image resampling filter.
type ResampleFilter struct {
	Support float64
	Kernel  func(float64) float64
}

// ConvolveOptions specifies options for convolution operations.
type ConvolveOptions struct {
	Normalize bool
	Abs       bool
	Bias      int
}

// DecodeOption is a functional option for Decode/Open.
type DecodeOption func(*decodeConfig)

// EncodeOption is a functional option for Encode/Save.
type EncodeOption func(*encodeConfig)

// ErrUnsupportedFormat is returned when the image format is not supported.
var ErrUnsupportedFormat = errors.New("imaging: unsupported image format")

// Resample filters — identical kernels to disintegration/imaging v1.6.2.
var (
	NearestNeighbor   = ResampleFilter{0, func(x float64) float64 { return 0 }}
	Box               = ResampleFilter{0.5, boxKernel}
	Linear            = ResampleFilter{1.0, linearKernel}
	Hermite           = ResampleFilter{1.0, hermiteKernel}
	MitchellNetravali = ResampleFilter{2.0, mitchellNetravaliKernel}
	CatmullRom        = ResampleFilter{2.0, catmullRomKernel}
	BSpline           = ResampleFilter{2.0, bSplineKernel}
	Gaussian          = ResampleFilter{2.0, gaussianKernel}
	Bartlett          = ResampleFilter{1.0, bartlettKernel}
	Lanczos           = ResampleFilter{3.0, lanczosKernel}
	Hann              = ResampleFilter{3.0, hannKernel}
	Hamming           = ResampleFilter{3.0, hammingKernel}
	Blackman          = ResampleFilter{3.0, blackmanKernel}
	Welch             = ResampleFilter{3.0, welchKernel}
	Cosine            = ResampleFilter{3.0, cosineKernel}
)

func boxKernel(x float64) float64 {
	x = math.Abs(x)
	if x <= 0.5 {
		return 1
	}
	return 0
}

func linearKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 1 {
		return 1 - x
	}
	return 0
}

func hermiteKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 1 {
		return (2*x-3)*x*x + 1
	}
	return 0
}

func mitchellNetravaliKernel(x float64) float64 {
	const b, c = 1.0 / 3.0, 1.0 / 3.0
	return bcSplineKernel(x, b, c)
}

func catmullRomKernel(x float64) float64 {
	return bcSplineKernel(x, 0, 0.5)
}

func bSplineKernel(x float64) float64 {
	return bcSplineKernel(x, 1, 0)
}

func bcSplineKernel(x, b, c float64) float64 {
	x = math.Abs(x)
	if x < 1 {
		return ((12-9*b-6*c)*x*x*x + (-18+12*b+6*c)*x*x + (6 - 2*b)) / 6
	}
	if x < 2 {
		return ((-b-6*c)*x*x*x + (6*b+30*c)*x*x + (-12*b-48*c)*x + (8*b + 24*c)) / 6
	}
	return 0
}

func gaussianKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 2 {
		return math.Exp(-2 * x * x) * math.Sqrt(2.0/math.Pi)
	}
	return 0
}

func bartlettKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 1 {
		return 1 - x
	}
	return 0
}

func sinc(x float64) float64 {
	if x == 0 {
		return 1
	}
	return math.Sin(math.Pi*x) / (math.Pi * x)
}

func lanczosKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * sinc(x/3)
	}
	return 0
}

func hannKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * (0.5 + 0.5*math.Cos(math.Pi*x/3))
	}
	return 0
}

func hammingKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * (0.54 + 0.46*math.Cos(math.Pi*x/3))
	}
	return 0
}

func blackmanKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * (0.42 - 0.5*math.Cos(math.Pi*x/3+math.Pi) + 0.08*math.Cos(2*math.Pi*x/3))
	}
	return 0
}

func welchKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * (1 - x*x/9)
	}
	return 0
}

func cosineKernel(x float64) float64 {
	x = math.Abs(x)
	if x < 3 {
		return sinc(x) * math.Cos(math.Pi*x/6)
	}
	return 0
}

// New creates a new image of the given size filled with the given color.
func New(width, height int, fillColor color.Color) *image.NRGBA {
	if width <= 0 || height <= 0 {
		return &image.NRGBA{}
	}

	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	c := color.NRGBAModel.Convert(fillColor).(color.NRGBA)

	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		return dst // already zeroed
	}

	// Fill pixel data
	for i := 0; i < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = c.R
		dst.Pix[i+1] = c.G
		dst.Pix[i+2] = c.B
		dst.Pix[i+3] = c.A
	}
	return dst
}

// Clone returns a copy of the given image as *image.NRGBA.
func Clone(img image.Image) *image.NRGBA {
	if img == nil {
		return &image.NRGBA{}
	}

	bounds := img.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	s := newScanner(img)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		dstY := y - bounds.Min.Y
		i := dstY * dst.Stride
		s.scan(bounds.Min.X, y, bounds.Max.X, y+1, dst.Pix[i:i+dst.Stride])
	}

	return dst
}

// anchorPt calculates the position of a rectangle of the given size
// within a parent rectangle based on the specified anchor point.
func anchorPt(parent image.Rectangle, w, h int, anchor Anchor) image.Point {
	px, py := parent.Dx(), parent.Dy()
	switch anchor {
	case TopLeft:
		return parent.Min
	case Top:
		return image.Pt(parent.Min.X+(px-w)/2, parent.Min.Y)
	case TopRight:
		return image.Pt(parent.Max.X-w, parent.Min.Y)
	case Left:
		return image.Pt(parent.Min.X, parent.Min.Y+(py-h)/2)
	case Right:
		return image.Pt(parent.Max.X-w, parent.Min.Y+(py-h)/2)
	case BottomLeft:
		return image.Pt(parent.Min.X, parent.Max.Y-h)
	case Bottom:
		return image.Pt(parent.Min.X+(px-w)/2, parent.Max.Y-h)
	case BottomRight:
		return image.Pt(parent.Max.X-w, parent.Max.Y-h)
	default: // Center
		return image.Pt(parent.Min.X+(px-w)/2, parent.Min.Y+(py-h)/2)
	}
}
