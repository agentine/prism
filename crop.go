package prism

import "image"

// Crop cuts out a rectangular region from the image.
func Crop(img image.Image, rect image.Rectangle) *image.NRGBA {
	bounds := img.Bounds()
	// Intersect with image bounds for safety.
	rect = rect.Intersect(bounds)
	if rect.Empty() {
		return &image.NRGBA{}
	}

	w := rect.Dx()
	h := rect.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	s := newScanner(img)

	parallel(0, h, func(y int) {
		dstOff := y * dst.Stride
		s.scan(rect.Min.X, rect.Min.Y+y, rect.Max.X, rect.Min.Y+y+1, dst.Pix[dstOff:dstOff+w*4])
	})

	return dst
}

// CropCenter cuts out a rectangle of the given size from the center of the image.
func CropCenter(img image.Image, width, height int) *image.NRGBA {
	return CropAnchor(img, width, height, Center)
}

// CropAnchor cuts out a rectangle of the given size from the image, anchored
// at the specified anchor point.
func CropAnchor(img image.Image, width, height int, anchor Anchor) *image.NRGBA {
	bounds := img.Bounds()
	pt := anchorPt(bounds, width, height, anchor)
	rect := image.Rect(pt.X, pt.Y, pt.X+width, pt.Y+height)
	return Crop(img, rect)
}
