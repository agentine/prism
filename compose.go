package prism

import "image"

// Paste pastes img over the background image at the given position.
// The paste area is safely intersected with the background bounds to
// prevent panics (fix for imaging #163).
func Paste(background, img image.Image, pos image.Point) *image.NRGBA {
	dst := Clone(background)
	src := Clone(img)

	dstBounds := dst.Bounds()
	srcBounds := src.Bounds()

	// Calculate the paste rectangle in destination coordinates.
	pasteRect := image.Rect(
		pos.X, pos.Y,
		pos.X+srcBounds.Dx(), pos.Y+srcBounds.Dy(),
	)

	// Intersect with destination bounds to prevent out-of-bounds writes.
	pasteRect = pasteRect.Intersect(dstBounds)
	if pasteRect.Empty() {
		return dst
	}

	for y := pasteRect.Min.Y; y < pasteRect.Max.Y; y++ {
		for x := pasteRect.Min.X; x < pasteRect.Max.X; x++ {
			srcX := x - pos.X
			srcY := y - pos.Y
			if srcX < 0 || srcY < 0 || srcX >= srcBounds.Dx() || srcY >= srcBounds.Dy() {
				continue
			}
			si := srcY*src.Stride + srcX*4
			di := y*dst.Stride + x*4
			dst.Pix[di+0] = src.Pix[si+0]
			dst.Pix[di+1] = src.Pix[si+1]
			dst.Pix[di+2] = src.Pix[si+2]
			dst.Pix[di+3] = src.Pix[si+3]
		}
	}

	return dst
}

// PasteCenter pastes img at the center of the background image.
func PasteCenter(background, img image.Image) *image.NRGBA {
	bgBounds := background.Bounds()
	imgBounds := img.Bounds()
	pos := image.Pt(
		(bgBounds.Dx()-imgBounds.Dx())/2,
		(bgBounds.Dy()-imgBounds.Dy())/2,
	)
	return Paste(background, img, pos)
}

// Overlay draws img over the background at the given position with the
// specified opacity (0.0 to 1.0). Alpha blending is applied correctly.
func Overlay(background, img image.Image, pos image.Point, opacity float64) *image.NRGBA {
	if opacity <= 0 {
		return Clone(background)
	}
	if opacity > 1 {
		opacity = 1
	}

	dst := Clone(background)
	src := Clone(img)

	dstBounds := dst.Bounds()
	srcBounds := src.Bounds()

	pasteRect := image.Rect(
		pos.X, pos.Y,
		pos.X+srcBounds.Dx(), pos.Y+srcBounds.Dy(),
	)
	pasteRect = pasteRect.Intersect(dstBounds)
	if pasteRect.Empty() {
		return dst
	}

	for y := pasteRect.Min.Y; y < pasteRect.Max.Y; y++ {
		for x := pasteRect.Min.X; x < pasteRect.Max.X; x++ {
			srcX := x - pos.X
			srcY := y - pos.Y
			if srcX < 0 || srcY < 0 || srcX >= srcBounds.Dx() || srcY >= srcBounds.Dy() {
				continue
			}
			si := srcY*src.Stride + srcX*4
			di := y*dst.Stride + x*4

			// Source alpha modified by opacity.
			srcA := float64(src.Pix[si+3]) / 255.0 * opacity
			dstA := float64(dst.Pix[di+3]) / 255.0

			// Porter-Duff over composite.
			outA := srcA + dstA*(1-srcA)
			if outA > 0 {
				dst.Pix[di+0] = clamp((float64(src.Pix[si+0])*srcA + float64(dst.Pix[di+0])*dstA*(1-srcA)) / outA)
				dst.Pix[di+1] = clamp((float64(src.Pix[si+1])*srcA + float64(dst.Pix[di+1])*dstA*(1-srcA)) / outA)
				dst.Pix[di+2] = clamp((float64(src.Pix[si+2])*srcA + float64(dst.Pix[di+2])*dstA*(1-srcA)) / outA)
				dst.Pix[di+3] = clamp(outA * 255)
			}
		}
	}

	return dst
}

// OverlayCenter draws img at the center of the background with the given opacity.
func OverlayCenter(background, img image.Image, opacity float64) *image.NRGBA {
	bgBounds := background.Bounds()
	imgBounds := img.Bounds()
	pos := image.Pt(
		(bgBounds.Dx()-imgBounds.Dx())/2,
		(bgBounds.Dy()-imgBounds.Dy())/2,
	)
	return Overlay(background, img, pos, opacity)
}
