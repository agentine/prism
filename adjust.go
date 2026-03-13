package prism

import (
	"image"
	"image/color"
	"math"
)

// AdjustBrightness changes the brightness of the image.
// percentage must be in the range [-100, 100].
func AdjustBrightness(img image.Image, percentage float64) *image.NRGBA {
	percentage = math.Max(-100, math.Min(100, percentage))
	shift := 255.0 * percentage / 100.0

	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = clamp(float64(i) + shift)
	}

	return applyLUT(img, lut)
}

// AdjustContrast changes the contrast of the image.
// percentage must be in the range [-100, 100].
func AdjustContrast(img image.Image, percentage float64) *image.NRGBA {
	percentage = math.Max(-100, math.Min(100, percentage))
	v := (100 + percentage) / 100

	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = clamp((float64(i)/255.0-0.5)*v*255.0 + 128)
	}

	return applyLUT(img, lut)
}

// AdjustGamma performs gamma correction on the image.
func AdjustGamma(img image.Image, gamma float64) *image.NRGBA {
	if gamma <= 0 {
		return Clone(img)
	}

	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = clamp(math.Pow(float64(i)/255.0, gamma) * 255.0)
	}

	return applyLUT(img, lut)
}

// AdjustSigmoid adjusts the contrast using a sigmoidal function.
// midpoint is the center of the transition (0-1), factor controls steepness.
func AdjustSigmoid(img image.Image, midpoint, factor float64) *image.NRGBA {
	a := math.Min(math.Max(midpoint, 0), 1)
	b := math.Abs(factor)
	sigmoid := func(x float64) float64 {
		return 1.0 / (1.0 + math.Exp(b*(a-x)))
	}

	sig0 := sigmoid(0)
	sig1 := sigmoid(1)
	e := sig1 - sig0

	var lut [256]uint8
	for i := 0; i < 256; i++ {
		x := float64(i) / 255.0
		if factor >= 0 {
			x = (sigmoid(x) - sig0) / e
		} else {
			x = -1.0 / b * math.Log(1.0/(x*e+sig0)-1.0) + a
		}
		lut[i] = clamp(x * 255.0)
	}

	return applyLUT(img, lut)
}

// AdjustSaturation changes the saturation of the image.
// percentage must be in the range [-100, 500].
func AdjustSaturation(img image.Image, percentage float64) *image.NRGBA {
	percentage = math.Max(-100, math.Min(500, percentage))
	factor := 1 + percentage/100

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
			si := x * 4
			r, g, b := float64(row[si]), float64(row[si+1]), float64(row[si+2])
			h, sat, l := rgbToHSL(r/255, g/255, b/255)
			sat *= factor
			sat = math.Max(0, math.Min(1, sat))
			rr, gg, bb := hslToRGB(h, sat, l)
			dst.Pix[dstOff+x*4+0] = clamp(rr * 255)
			dst.Pix[dstOff+x*4+1] = clamp(gg * 255)
			dst.Pix[dstOff+x*4+2] = clamp(bb * 255)
			dst.Pix[dstOff+x*4+3] = row[si+3]
		}
	})

	return dst
}

// AdjustHue shifts the hue of the image by the given amount (in degrees).
func AdjustHue(img image.Image, shift float64) *image.NRGBA {
	shift = math.Mod(shift, 360)
	if shift < 0 {
		shift += 360
	}
	hueShift := shift / 360

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
			si := x * 4
			r, g, b := float64(row[si])/255, float64(row[si+1])/255, float64(row[si+2])/255
			h, sat, l := rgbToHSL(r, g, b)
			h += hueShift
			if h > 1 {
				h -= 1
			}
			rr, gg, bb := hslToRGB(h, sat, l)
			dst.Pix[dstOff+x*4+0] = clamp(rr * 255)
			dst.Pix[dstOff+x*4+1] = clamp(gg * 255)
			dst.Pix[dstOff+x*4+2] = clamp(bb * 255)
			dst.Pix[dstOff+x*4+3] = row[si+3]
		}
	})

	return dst
}

// AdjustFunc applies a custom per-pixel color transformation.
func AdjustFunc(img image.Image, fn func(c color.NRGBA) color.NRGBA) *image.NRGBA {
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
			si := x * 4
			c := fn(color.NRGBA{R: row[si], G: row[si+1], B: row[si+2], A: row[si+3]})
			dst.Pix[dstOff+x*4+0] = c.R
			dst.Pix[dstOff+x*4+1] = c.G
			dst.Pix[dstOff+x*4+2] = c.B
			dst.Pix[dstOff+x*4+3] = c.A
		}
	})

	return dst
}

// Grayscale converts the image to grayscale.
func Grayscale(img image.Image) *image.NRGBA {
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
			si := x * 4
			// ITU-R BT.601 luminance coefficients.
			gray := clamp(0.299*float64(row[si]) + 0.587*float64(row[si+1]) + 0.114*float64(row[si+2]))
			dst.Pix[dstOff+x*4+0] = gray
			dst.Pix[dstOff+x*4+1] = gray
			dst.Pix[dstOff+x*4+2] = gray
			dst.Pix[dstOff+x*4+3] = row[si+3]
		}
	})

	return dst
}

// Invert inverts the colors of the image.
func Invert(img image.Image) *image.NRGBA {
	var lut [256]uint8
	for i := 0; i < 256; i++ {
		lut[i] = uint8(255 - i)
	}
	return applyLUT(img, lut)
}

// applyLUT applies a lookup table to the R, G, B channels of each pixel.
func applyLUT(img image.Image, lut [256]uint8) *image.NRGBA {
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
			si := x * 4
			dst.Pix[dstOff+x*4+0] = lut[row[si+0]]
			dst.Pix[dstOff+x*4+1] = lut[row[si+1]]
			dst.Pix[dstOff+x*4+2] = lut[row[si+2]]
			dst.Pix[dstOff+x*4+3] = row[si+3] // alpha unchanged
		}
	})

	return dst
}

// HSL conversion helpers.

func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l = (max + min) / 2

	if max == min {
		return 0, 0, l
	}

	d := max - min
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}

	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6
	return
}

func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)
	return
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}
