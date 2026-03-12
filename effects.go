package prism

import (
	"image"
	"math"
)

// Blur applies a Gaussian blur to the image.
// sigma is the standard deviation of the Gaussian kernel; larger values
// produce a stronger blur.
func Blur(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	// Build 1D Gaussian kernel.
	radius := int(math.Ceil(sigma * 3))
	if radius < 1 {
		radius = 1
	}
	size := 2*radius + 1
	kernel := make([]float64, size)
	var sum float64
	for i := -radius; i <= radius; i++ {
		v := math.Exp(-float64(i*i) / (2 * sigma * sigma))
		kernel[i+radius] = v
		sum += v
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	// Two-pass separable convolution: horizontal then vertical.
	tmp := blurHorizontal(img, kernel, radius)
	return blurVertical(tmp, kernel, radius)
}

// Sharpen sharpens the image using unsharp masking.
// sigma controls the blur radius; a larger sigma produces stronger sharpening.
func Sharpen(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	blurred := Blur(img, sigma)
	src := Clone(img)

	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))

	parallel(0, h, func(y int) {
		for x := 0; x < w; x++ {
			si := y*src.Stride + x*4
			bi := y*blurred.Stride + x*4
			di := y*dst.Stride + x*4

			for c := 0; c < 3; c++ {
				// Unsharp mask: original + (original - blurred)
				v := 2*float64(src.Pix[si+c]) - float64(blurred.Pix[bi+c])
				dst.Pix[di+c] = clamp(v)
			}
			dst.Pix[di+3] = src.Pix[si+3]
		}
	})

	return dst
}

func blurHorizontal(img image.Image, kernel []float64, radius int) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		dstOff := y * dst.Stride

		for x := 0; x < w; x++ {
			var r, g, b, a float64
			for i := -radius; i <= radius; i++ {
				sx := x + i
				if sx < 0 {
					sx = 0
				} else if sx >= w {
					sx = w - 1
				}
				si := sx * 4
				wt := kernel[i+radius]
				r += float64(row[si+0]) * wt
				g += float64(row[si+1]) * wt
				b += float64(row[si+2]) * wt
				a += float64(row[si+3]) * wt
			}
			dst.Pix[dstOff+x*4+0] = clamp(r)
			dst.Pix[dstOff+x*4+1] = clamp(g)
			dst.Pix[dstOff+x*4+2] = clamp(b)
			dst.Pix[dstOff+x*4+3] = clamp(a)
		}
	})

	return dst
}

func blurVertical(img *image.NRGBA, kernel []float64, radius int) *image.NRGBA {
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))

	parallel(0, h, func(y int) {
		dstOff := y * dst.Stride
		for x := 0; x < w; x++ {
			var r, g, b, a float64
			for i := -radius; i <= radius; i++ {
				sy := y + i
				if sy < 0 {
					sy = 0
				} else if sy >= h {
					sy = h - 1
				}
				si := sy*img.Stride + x*4
				wt := kernel[i+radius]
				r += float64(img.Pix[si+0]) * wt
				g += float64(img.Pix[si+1]) * wt
				b += float64(img.Pix[si+2]) * wt
				a += float64(img.Pix[si+3]) * wt
			}
			dst.Pix[dstOff+x*4+0] = clamp(r)
			dst.Pix[dstOff+x*4+1] = clamp(g)
			dst.Pix[dstOff+x*4+2] = clamp(b)
			dst.Pix[dstOff+x*4+3] = clamp(a)
		}
	})

	return dst
}
