package prism

import "image"

// Histogram returns the normalized luminance histogram of the image.
// Each entry is the fraction of pixels at that luminance level (0-255).
func Histogram(img image.Image) [256]float64 {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	total := w * h

	if total == 0 {
		return [256]float64{}
	}

	var counts [256]int
	s := newScanner(img)

	for y := 0; y < h; y++ {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		for x := 0; x < w; x++ {
			si := x * 4
			// ITU-R BT.601 luminance.
			lum := 0.299*float64(row[si]) + 0.587*float64(row[si+1]) + 0.114*float64(row[si+2])
			idx := int(lum + 0.5)
			if idx > 255 {
				idx = 255
			}
			counts[idx]++
		}
	}

	var hist [256]float64
	ft := float64(total)
	for i, c := range counts {
		hist[i] = float64(c) / ft
	}

	return hist
}
