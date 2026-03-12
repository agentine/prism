package prism

import (
	"image"
	"image/color"
)

// scanner is a bounds-safe pixel scanner that converts any image.Image
// to NRGBA pixel data. All pixel access is bounds-checked to prevent
// index out of bounds panics (CVE-2023-36308 fix).
type scanner struct {
	img    image.Image
	bounds image.Rectangle
}

func newScanner(img image.Image) *scanner {
	return &scanner{
		img:    img,
		bounds: img.Bounds(),
	}
}

// scan reads pixels from the source image within the specified rectangle
// and writes them as NRGBA (4 bytes per pixel) into dst.
// The rectangle is clamped to the image bounds to prevent out-of-bounds access.
func (s *scanner) scan(x1, y1, x2, y2 int, dst []byte) {
	// Clamp to image bounds for safety (CVE-2023-36308).
	if x1 < s.bounds.Min.X {
		x1 = s.bounds.Min.X
	}
	if y1 < s.bounds.Min.Y {
		y1 = s.bounds.Min.Y
	}
	if x2 > s.bounds.Max.X {
		x2 = s.bounds.Max.X
	}
	if y2 > s.bounds.Max.Y {
		y2 = s.bounds.Max.Y
	}

	width := x2 - x1
	if width <= 0 || y2 <= y1 {
		return
	}

	// Fast path for known image types.
	switch src := s.img.(type) {
	case *image.NRGBA:
		s.scanNRGBA(src, x1, y1, x2, y2, dst)
	case *image.RGBA:
		s.scanRGBA(src, x1, y1, x2, y2, dst)
	case *image.Gray:
		s.scanGray(src, x1, y1, x2, y2, dst)
	default:
		s.scanGeneric(x1, y1, x2, y2, dst)
	}
}

func (s *scanner) scanNRGBA(src *image.NRGBA, x1, y1, x2, y2 int, dst []byte) {
	width := x2 - x1
	for y := y1; y < y2; y++ {
		dstOff := (y - y1) * width * 4
		srcOff := (y-src.Rect.Min.Y)*src.Stride + (x1-src.Rect.Min.X)*4
		srcEnd := srcOff + width*4

		// Bounds check: ensure srcOff and srcEnd are within src.Pix.
		if srcOff < 0 || srcEnd > len(src.Pix) || srcOff >= srcEnd {
			continue
		}
		dstEnd := dstOff + width*4
		if dstEnd > len(dst) {
			continue
		}

		copy(dst[dstOff:dstEnd], src.Pix[srcOff:srcEnd])
	}
}

func (s *scanner) scanRGBA(src *image.RGBA, x1, y1, x2, y2 int, dst []byte) {
	width := x2 - x1
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			srcOff := (y-src.Rect.Min.Y)*src.Stride + (x-src.Rect.Min.X)*4
			if srcOff < 0 || srcOff+3 >= len(src.Pix) {
				continue
			}
			r, g, b, a := src.Pix[srcOff], src.Pix[srcOff+1], src.Pix[srcOff+2], src.Pix[srcOff+3]

			dstOff := ((y - y1) * width * 4) + (x-x1)*4
			if dstOff+3 >= len(dst) {
				continue
			}

			// Convert premultiplied alpha to non-premultiplied.
			if a == 0 {
				dst[dstOff] = 0
				dst[dstOff+1] = 0
				dst[dstOff+2] = 0
				dst[dstOff+3] = 0
			} else if a == 255 {
				dst[dstOff] = r
				dst[dstOff+1] = g
				dst[dstOff+2] = b
				dst[dstOff+3] = 255
			} else {
				dst[dstOff] = uint8(uint16(r) * 255 / uint16(a))
				dst[dstOff+1] = uint8(uint16(g) * 255 / uint16(a))
				dst[dstOff+2] = uint8(uint16(b) * 255 / uint16(a))
				dst[dstOff+3] = a
			}
		}
	}
}

func (s *scanner) scanGray(src *image.Gray, x1, y1, x2, y2 int, dst []byte) {
	width := x2 - x1
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			srcOff := (y-src.Rect.Min.Y)*src.Stride + (x - src.Rect.Min.X)
			if srcOff < 0 || srcOff >= len(src.Pix) {
				continue
			}
			gray := src.Pix[srcOff]

			dstOff := ((y - y1) * width * 4) + (x-x1)*4
			if dstOff+3 >= len(dst) {
				continue
			}

			dst[dstOff] = gray
			dst[dstOff+1] = gray
			dst[dstOff+2] = gray
			dst[dstOff+3] = 255
		}
	}
}

func (s *scanner) scanGeneric(x1, y1, x2, y2 int, dst []byte) {
	width := x2 - x1
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			dstOff := ((y - y1) * width * 4) + (x-x1)*4
			if dstOff+3 >= len(dst) {
				continue
			}

			c := s.img.At(x, y)
			nc := color.NRGBAModel.Convert(c).(color.NRGBA)
			dst[dstOff] = nc.R
			dst[dstOff+1] = nc.G
			dst[dstOff+2] = nc.B
			dst[dstOff+3] = nc.A
		}
	}
}
