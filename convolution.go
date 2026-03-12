package prism

import (
	"image"
	"math"
)

// Convolve3x3 applies a 3x3 convolution kernel to the image.
func Convolve3x3(img image.Image, kernel [9]float64, options *ConvolveOptions) *image.NRGBA {
	return convolve(img, kernel[:], 3, options)
}

// Convolve5x5 applies a 5x5 convolution kernel to the image.
func Convolve5x5(img image.Image, kernel [25]float64, options *ConvolveOptions) *image.NRGBA {
	return convolve(img, kernel[:], 5, options)
}

func convolve(img image.Image, kernel []float64, size int, options *ConvolveOptions) *image.NRGBA {
	if options == nil {
		options = &ConvolveOptions{}
	}

	// Normalize kernel if requested.
	k := make([]float64, len(kernel))
	copy(k, kernel)

	if options.Normalize {
		var sum float64
		for _, v := range k {
			sum += v
		}
		if sum != 0 {
			for i := range k {
				k[i] /= sum
			}
		}
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := newScanner(img)

	// Read entire image into buffer for random access.
	src := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, src[y*w*4:(y+1)*w*4])
	}

	half := size / 2
	bias := float64(options.Bias)

	parallel(0, h, func(y int) {
		dstOff := y * dst.Stride
		for x := 0; x < w; x++ {
			var r, g, b float64

			ki := 0
			for ky := -half; ky <= half; ky++ {
				for kx := -half; kx <= half; kx++ {
					sx := x + kx
					sy := y + ky
					if sx < 0 {
						sx = 0
					} else if sx >= w {
						sx = w - 1
					}
					if sy < 0 {
						sy = 0
					} else if sy >= h {
						sy = h - 1
					}

					si := sy*w*4 + sx*4
					wt := k[ki]
					r += float64(src[si+0]) * wt
					g += float64(src[si+1]) * wt
					b += float64(src[si+2]) * wt
					ki++
				}
			}

			if options.Abs {
				r = math.Abs(r)
				g = math.Abs(g)
				b = math.Abs(b)
			}

			r += bias
			g += bias
			b += bias

			// Alpha from center pixel.
			centerIdx := y*w*4 + x*4

			dst.Pix[dstOff+x*4+0] = clamp(r)
			dst.Pix[dstOff+x*4+1] = clamp(g)
			dst.Pix[dstOff+x*4+2] = clamp(b)
			dst.Pix[dstOff+x*4+3] = src[centerIdx+3]
		}
	})

	return dst
}
