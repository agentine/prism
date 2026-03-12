package prism

import (
	"image"
	"image/color"
	"math"
)

// FlipH returns a horizontally flipped copy of the image.
func FlipH(img image.Image) *image.NRGBA {
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
			srcX := (w - 1 - x) * 4
			dst.Pix[dstOff+x*4+0] = row[srcX+0]
			dst.Pix[dstOff+x*4+1] = row[srcX+1]
			dst.Pix[dstOff+x*4+2] = row[srcX+2]
			dst.Pix[dstOff+x*4+3] = row[srcX+3]
		}
	})

	return dst
}

// FlipV returns a vertically flipped copy of the image.
func FlipV(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		srcY := h - 1 - y
		dstOff := y * dst.Stride
		s.scan(bounds.Min.X, bounds.Min.Y+srcY, bounds.Max.X, bounds.Min.Y+srcY+1, dst.Pix[dstOff:dstOff+w*4])
	})

	return dst
}

// Rotate90 rotates the image 90 degrees clockwise.
func Rotate90(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		// src(x, y) -> dst(h-1-y, x) => dst col = h-1-y, dst row = x
		dstX := h - 1 - y
		for x := 0; x < w; x++ {
			si := x * 4
			di := x*dst.Stride + dstX*4
			dst.Pix[di+0] = row[si+0]
			dst.Pix[di+1] = row[si+1]
			dst.Pix[di+2] = row[si+2]
			dst.Pix[di+3] = row[si+3]
		}
	})

	return dst
}

// Rotate180 rotates the image 180 degrees.
func Rotate180(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		dstY := h - 1 - y
		dstOff := dstY * dst.Stride
		for x := 0; x < w; x++ {
			srcX := (w - 1 - x) * 4
			dst.Pix[dstOff+x*4+0] = row[srcX+0]
			dst.Pix[dstOff+x*4+1] = row[srcX+1]
			dst.Pix[dstOff+x*4+2] = row[srcX+2]
			dst.Pix[dstOff+x*4+3] = row[srcX+3]
		}
	})

	return dst
}

// Rotate270 rotates the image 270 degrees clockwise (90 degrees counter-clockwise).
func Rotate270(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		// src(x, y) -> dst(y, w-1-x) => dst col = y, dst row = w-1-x
		dstX := y
		for x := 0; x < w; x++ {
			si := x * 4
			dstY := w - 1 - x
			di := dstY*dst.Stride + dstX*4
			dst.Pix[di+0] = row[si+0]
			dst.Pix[di+1] = row[si+1]
			dst.Pix[di+2] = row[si+2]
			dst.Pix[di+3] = row[si+3]
		}
	})

	return dst
}

// Transpose flips the image along the top-left to bottom-right diagonal.
// Equivalent to Rotate90 then FlipH.
func Transpose(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		// src(x, y) -> dst(y, x)
		dstX := y
		for x := 0; x < w; x++ {
			si := x * 4
			di := x*dst.Stride + dstX*4
			dst.Pix[di+0] = row[si+0]
			dst.Pix[di+1] = row[si+1]
			dst.Pix[di+2] = row[si+2]
			dst.Pix[di+3] = row[si+3]
		}
	})

	return dst
}

// Transverse flips the image along the top-right to bottom-left diagonal.
// Equivalent to Rotate90 then FlipV.
func Transverse(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, h, w))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		row := make([]byte, w*4)
		s.scan(bounds.Min.X, bounds.Min.Y+y, bounds.Max.X, bounds.Min.Y+y+1, row)
		// src(x, y) -> dst(h-1-y, w-1-x)
		dstX := h - 1 - y
		for x := 0; x < w; x++ {
			si := x * 4
			dstY := w - 1 - x
			di := dstY*dst.Stride + dstX*4
			dst.Pix[di+0] = row[si+0]
			dst.Pix[di+1] = row[si+1]
			dst.Pix[di+2] = row[si+2]
			dst.Pix[di+3] = row[si+3]
		}
	})

	return dst
}

// Rotate rotates the image by the given angle (in degrees, counter-clockwise)
// around its center. Empty areas are filled with bgColor.
func Rotate(img image.Image, angle float64, bgColor color.Color) *image.NRGBA {
	// Normalize angle to [0, 360).
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}

	// Handle exact 90-degree rotations without interpolation.
	switch angle {
	case 0:
		return Clone(img)
	case 90:
		return Rotate90(img)
	case 180:
		return Rotate180(img)
	case 270:
		return Rotate270(img)
	}

	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Calculate new bounds after rotation.
	rad := angle * math.Pi / 180
	sinA := math.Abs(math.Sin(rad))
	cosA := math.Abs(math.Cos(rad))

	dstW := int(math.Ceil(float64(srcW)*cosA + float64(srcH)*sinA))
	dstH := int(math.Ceil(float64(srcW)*sinA + float64(srcH)*cosA))

	if dstW <= 0 {
		dstW = 1
	}
	if dstH <= 0 {
		dstH = 1
	}

	bg := color.NRGBAModel.Convert(bgColor).(color.NRGBA)
	dst := New(dstW, dstH, bg)
	s := newScanner(img)

	// Center of source and destination.
	srcCX := float64(srcW) / 2
	srcCY := float64(srcH) / 2
	dstCX := float64(dstW) / 2
	dstCY := float64(dstH) / 2

	sinR := math.Sin(-rad)
	cosR := math.Cos(-rad)

	parallel(0, dstH, func(y int) {
		for x := 0; x < dstW; x++ {
			// Map destination pixel to source pixel (inverse rotation).
			dx := float64(x) - dstCX + 0.5
			dy := float64(y) - dstCY + 0.5
			srcX := dx*cosR - dy*sinR + srcCX
			srcY := dx*sinR + dy*cosR + srcCY

			// Bilinear interpolation.
			sx := int(math.Floor(srcX))
			sy := int(math.Floor(srcY))

			if sx < 0 || sy < 0 || sx >= srcW-1 || sy >= srcH-1 {
				// Out of bounds — use background (already filled).
				if sx >= 0 && sy >= 0 && sx < srcW && sy < srcH {
					// Edge pixel — use nearest.
					pix := make([]byte, 4)
					s.scan(bounds.Min.X+sx, bounds.Min.Y+sy, bounds.Min.X+sx+1, bounds.Min.Y+sy+1, pix)
					di := y*dst.Stride + x*4
					dst.Pix[di+0] = pix[0]
					dst.Pix[di+1] = pix[1]
					dst.Pix[di+2] = pix[2]
					dst.Pix[di+3] = pix[3]
				}
				continue
			}

			fx := srcX - float64(sx)
			fy := srcY - float64(sy)

			// Read 2x2 block.
			block := make([]byte, 2*4)
			row1 := make([]byte, 2*4)
			s.scan(bounds.Min.X+sx, bounds.Min.Y+sy, bounds.Min.X+sx+2, bounds.Min.Y+sy+1, block)
			s.scan(bounds.Min.X+sx, bounds.Min.Y+sy+1, bounds.Min.X+sx+2, bounds.Min.Y+sy+2, row1)

			r00, g00, b00, a00 := float64(block[0]), float64(block[1]), float64(block[2]), float64(block[3])
			r10, g10, b10, a10 := float64(block[4]), float64(block[5]), float64(block[6]), float64(block[7])
			r01, g01, b01, a01 := float64(row1[0]), float64(row1[1]), float64(row1[2]), float64(row1[3])
			r11, g11, b11, a11 := float64(row1[4]), float64(row1[5]), float64(row1[6]), float64(row1[7])

			r := bilinear(r00, r10, r01, r11, fx, fy)
			g := bilinear(g00, g10, g01, g11, fx, fy)
			b := bilinear(b00, b10, b01, b11, fx, fy)
			a := bilinear(a00, a10, a01, a11, fx, fy)

			di := y*dst.Stride + x*4
			dst.Pix[di+0] = clamp(r)
			dst.Pix[di+1] = clamp(g)
			dst.Pix[di+2] = clamp(b)
			dst.Pix[di+3] = clamp(a)
		}
	})

	return dst
}

func bilinear(v00, v10, v01, v11, fx, fy float64) float64 {
	return v00*(1-fx)*(1-fy) + v10*fx*(1-fy) + v01*(1-fx)*fy + v11*fx*fy
}
