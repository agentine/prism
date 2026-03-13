package prism

import (
	"image"
	"math"
)

// Resize resizes the image to the specified width and height using the given
// resample filter. If one of width or height is 0, the image aspect ratio is
// preserved. If both are 0, the original image is returned as a copy.
func Resize(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW <= 0 || srcH <= 0 {
		return &image.NRGBA{}
	}

	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	if width == 0 && height == 0 {
		return Clone(img)
	}

	// Auto-calculate missing dimension to preserve aspect ratio.
	if width == 0 {
		width = int(math.Round(float64(height) * float64(srcW) / float64(srcH)))
		if width == 0 {
			width = 1
		}
	}
	if height == 0 {
		height = int(math.Round(float64(width) * float64(srcH) / float64(srcW)))
		if height == 0 {
			height = 1
		}
	}

	if int64(width)*int64(height) > maxDimPixels {
		return &image.NRGBA{}
	}

	if filter.Support <= 0 {
		// NearestNeighbor — no interpolation.
		return resizeNearest(img, width, height)
	}

	// Two-pass separable resize: horizontal then vertical.
	tmp := resizeHorizontal(img, width, filter)
	return resizeVertical(tmp, height, filter)
}

// Fit scales the image to fit within the given dimensions, preserving aspect
// ratio. The resulting image will be at most width x height.
func Fit(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA {
	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	if srcW <= 0 || srcH <= 0 || width <= 0 || height <= 0 {
		return &image.NRGBA{}
	}

	if srcW <= float64(width) && srcH <= float64(height) {
		return Clone(img)
	}

	scaleW := float64(width) / srcW
	scaleH := float64(height) / srcH
	scale := math.Min(scaleW, scaleH)

	newW := int(math.Round(srcW * scale))
	newH := int(math.Round(srcH * scale))

	if newW <= 0 {
		newW = 1
	}
	if newH <= 0 {
		newH = 1
	}

	return Resize(img, newW, newH, filter)
}

// Fill scales the image to fill the given dimensions, preserving aspect ratio,
// then crops to the exact size using the specified anchor point.
func Fill(img image.Image, width, height int, anchor Anchor, filter ResampleFilter) *image.NRGBA {
	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	if srcW <= 0 || srcH <= 0 || width <= 0 || height <= 0 {
		return &image.NRGBA{}
	}

	scaleW := float64(width) / srcW
	scaleH := float64(height) / srcH
	scale := math.Max(scaleW, scaleH)

	newW := int(math.Round(srcW * scale))
	newH := int(math.Round(srcH * scale))

	if newW <= 0 {
		newW = 1
	}
	if newH <= 0 {
		newH = 1
	}

	resized := Resize(img, newW, newH, filter)
	return CropAnchor(resized, width, height, anchor)
}

// Thumbnail creates a thumbnail of the image. It is equivalent to
// Fill(img, width, height, Center, filter).
func Thumbnail(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA {
	return Fill(img, width, height, Center, filter)
}

// resizeNearest performs nearest-neighbor resizing.
func resizeNearest(img image.Image, width, height int) *image.NRGBA {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	s := newScanner(img)

	parallel(0, height, func(y int) {
		srcY := bounds.Min.Y + int(float64(y)*float64(srcH)/float64(height))
		if srcY >= bounds.Max.Y {
			srcY = bounds.Max.Y - 1
		}
		row := make([]byte, srcW*4)
		s.scan(bounds.Min.X, srcY, bounds.Max.X, srcY+1, row)

		dstOff := y * dst.Stride
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * float64(srcW) / float64(width))
			if srcX >= srcW {
				srcX = srcW - 1
			}
			si := srcX * 4
			dst.Pix[dstOff+x*4+0] = row[si+0]
			dst.Pix[dstOff+x*4+1] = row[si+1]
			dst.Pix[dstOff+x*4+2] = row[si+2]
			dst.Pix[dstOff+x*4+3] = row[si+3]
		}
	})

	return dst
}

// resizeHorizontal resizes the image horizontally using the filter from the
// first call. It reads pixels via scanner and writes to a new image.
func resizeHorizontal(img image.Image, width int, filter ResampleFilter) *image.NRGBA {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if width == srcW {
		return Clone(img)
	}

	dst := image.NewNRGBA(image.Rect(0, 0, width, srcH))
	weights := precomputeWeights(width, srcW, filter)
	s := newScanner(img)

	parallel(0, srcH, func(y int) {
		srcY := bounds.Min.Y + y
		row := make([]byte, srcW*4)
		s.scan(bounds.Min.X, srcY, bounds.Max.X, srcY+1, row)

		dstOff := y * dst.Stride
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			for _, w := range weights[x] {
				si := w.index * 4
				r += float64(row[si+0]) * w.weight
				g += float64(row[si+1]) * w.weight
				b += float64(row[si+2]) * w.weight
				a += float64(row[si+3]) * w.weight
			}
			dst.Pix[dstOff+x*4+0] = clamp(r)
			dst.Pix[dstOff+x*4+1] = clamp(g)
			dst.Pix[dstOff+x*4+2] = clamp(b)
			dst.Pix[dstOff+x*4+3] = clamp(a)
		}
	})

	return dst
}

// resizeVertical resizes the image vertically using the filter.
func resizeVertical(img *image.NRGBA, height int, filter ResampleFilter) *image.NRGBA {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()

	if height == srcH {
		return img
	}

	dst := image.NewNRGBA(image.Rect(0, 0, srcW, height))
	weights := precomputeWeights(height, srcH, filter)

	parallel(0, height, func(y int) {
		dstOff := y * dst.Stride
		for x := 0; x < srcW; x++ {
			var r, g, b, a float64
			for _, w := range weights[y] {
				si := w.index*img.Stride + x*4
				r += float64(img.Pix[si+0]) * w.weight
				g += float64(img.Pix[si+1]) * w.weight
				b += float64(img.Pix[si+2]) * w.weight
				a += float64(img.Pix[si+3]) * w.weight
			}
			dst.Pix[dstOff+x*4+0] = clamp(r)
			dst.Pix[dstOff+x*4+1] = clamp(g)
			dst.Pix[dstOff+x*4+2] = clamp(b)
			dst.Pix[dstOff+x*4+3] = clamp(a)
		}
	})

	return dst
}

type weightEntry struct {
	index  int
	weight float64
}

// precomputeWeights precomputes the filter weights for resampling from
// srcSize to dstSize pixels.
func precomputeWeights(dstSize, srcSize int, filter ResampleFilter) [][]weightEntry {
	ratio := float64(srcSize) / float64(dstSize)
	scale := math.Max(ratio, 1.0)
	support := filter.Support * scale

	weights := make([][]weightEntry, dstSize)

	for i := 0; i < dstSize; i++ {
		center := (float64(i)+0.5)*ratio - 0.5
		lo := maxInt(int(math.Floor(center-support)), 0)
		hi := minInt(int(math.Ceil(center+support)), srcSize-1)

		var entries []weightEntry
		var totalWeight float64

		for j := lo; j <= hi; j++ {
			w := filter.Kernel((float64(j) - center) / scale)
			if w != 0 {
				entries = append(entries, weightEntry{index: j, weight: w})
				totalWeight += w
			}
		}

		// Normalize weights so they sum to 1.
		if totalWeight != 0 {
			for k := range entries {
				entries[k].weight /= totalWeight
			}
		}

		weights[i] = entries
	}

	return weights
}
