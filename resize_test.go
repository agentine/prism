package prism

import (
	"image"
	"image/color"
	"testing"
)

func TestResize(t *testing.T) {
	src := New(100, 50, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	dst := Resize(src, 50, 25, Lanczos)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 25 {
		t.Fatalf("size: got %dx%d, want 50x25", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestResize_AutoWidth(t *testing.T) {
	src := New(100, 50, color.White)
	dst := Resize(src, 0, 25, Lanczos)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 25 {
		t.Fatalf("size: got %dx%d, want 50x25", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestResize_AutoHeight(t *testing.T) {
	src := New(100, 50, color.White)
	dst := Resize(src, 50, 0, Lanczos)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 25 {
		t.Fatalf("size: got %dx%d, want 50x25", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestResize_BothZero(t *testing.T) {
	src := New(100, 50, color.White)
	dst := Resize(src, 0, 0, Lanczos)
	if dst.Bounds().Dx() != 100 || dst.Bounds().Dy() != 50 {
		t.Fatalf("expected clone, got %dx%d", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestResize_NearestNeighbor(t *testing.T) {
	src := New(100, 100, color.NRGBA{R: 128, G: 64, B: 32, A: 255})
	dst := Resize(src, 50, 50, NearestNeighbor)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	c := dst.NRGBAAt(25, 25)
	if c.R != 128 || c.G != 64 || c.B != 32 || c.A != 255 {
		t.Fatalf("color mismatch: got %v", c)
	}
}

func TestFit(t *testing.T) {
	src := New(200, 100, color.White)
	dst := Fit(src, 50, 50, Lanczos)
	// Should fit within 50x50 preserving 2:1 ratio -> 50x25
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 25 {
		t.Fatalf("size: got %dx%d, want 50x25", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestFit_AlreadySmaller(t *testing.T) {
	src := New(10, 10, color.White)
	dst := Fit(src, 50, 50, Lanczos)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Fatalf("expected clone, got %dx%d", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestFill(t *testing.T) {
	src := New(200, 100, color.White)
	dst := Fill(src, 50, 50, Center, Lanczos)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestThumbnail(t *testing.T) {
	src := New(200, 100, color.White)
	dst := Thumbnail(src, 50, 50, Lanczos)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestResize_DifferentFiltersProduceDifferentOutput(t *testing.T) {
	// Create a test image with varying pixel values to ensure filters differ.
	src := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			src.SetNRGBA(x, y, color.NRGBA{
				R: uint8(x * 255 / 100),
				G: uint8(y * 255 / 100),
				B: uint8((x + y) * 255 / 200),
				A: 255,
			})
		}
	}

	resLanczos := Resize(src, 30, 30, Lanczos)
	resLinear := Resize(src, 30, 30, Linear)

	// Count differing pixels — different filters must produce different output.
	diffs := 0
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			c1 := resLanczos.NRGBAAt(x, y)
			c2 := resLinear.NRGBAAt(x, y)
			if c1 != c2 {
				diffs++
			}
		}
	}
	if diffs == 0 {
		t.Fatal("Lanczos and Linear filters produced identical output — filter parameter likely not passed through")
	}
}

func TestCrop(t *testing.T) {
	src := New(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	dst := Crop(src, image.Rect(10, 10, 60, 60))
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	c := dst.NRGBAAt(0, 0)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("color: got %v, want red", c)
	}
}

func TestCrop_OutOfBounds(t *testing.T) {
	src := New(100, 100, color.White)
	dst := Crop(src, image.Rect(-10, -10, 200, 200))
	// Intersected with bounds: 0,0,100,100
	if dst.Bounds().Dx() != 100 || dst.Bounds().Dy() != 100 {
		t.Fatalf("size: got %dx%d, want 100x100", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestCropCenter(t *testing.T) {
	src := New(100, 100, color.White)
	dst := CropCenter(src, 50, 50)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestCropAnchor(t *testing.T) {
	src := New(100, 100, color.White)
	dst := CropAnchor(src, 50, 50, TopLeft)
	if dst.Bounds().Dx() != 50 || dst.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 50x50", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}
