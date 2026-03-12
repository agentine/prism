package prism

import (
	"image/color"
	"testing"
)

func TestFlipH(t *testing.T) {
	src := New(10, 10, color.Black)
	// Set top-left pixel to red.
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := FlipH(src)
	// After horizontal flip, top-left -> top-right.
	c := dst.NRGBAAt(9, 0)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("FlipH: expected red at (9,0), got %v", c)
	}
	// Top-left should now be black.
	c2 := dst.NRGBAAt(0, 0)
	if c2.R != 0 {
		t.Fatalf("FlipH: expected black at (0,0), got %v", c2)
	}
}

func TestFlipV(t *testing.T) {
	src := New(10, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := FlipV(src)
	// After vertical flip, top-left -> bottom-left.
	c := dst.NRGBAAt(0, 9)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("FlipV: expected red at (0,9), got %v", c)
	}
}

func TestRotate90(t *testing.T) {
	src := New(20, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := Rotate90(src)
	// 20x10 -> 10x20
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 20 {
		t.Fatalf("Rotate90: got %dx%d, want 10x20", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	// src(0,0) -> dst(9,0) for 90° clockwise
	c := dst.NRGBAAt(9, 0)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("Rotate90: expected red at (9,0), got %v", c)
	}
}

func TestRotate90_RegressionPanic159(t *testing.T) {
	// Regression test for imaging #159 — Rotate90 panic on certain dimensions.
	sizes := [][2]int{{1, 1}, {1, 100}, {100, 1}, {3, 7}, {7, 3}, {1000, 1}}
	for _, sz := range sizes {
		src := New(sz[0], sz[1], color.White)
		dst := Rotate90(src) // must not panic
		if dst.Bounds().Dx() != sz[1] || dst.Bounds().Dy() != sz[0] {
			t.Fatalf("Rotate90(%dx%d): got %dx%d, want %dx%d",
				sz[0], sz[1], dst.Bounds().Dx(), dst.Bounds().Dy(), sz[1], sz[0])
		}
	}
}

func TestRotate180(t *testing.T) {
	src := New(10, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := Rotate180(src)
	// src(0,0) -> dst(9,9)
	c := dst.NRGBAAt(9, 9)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("Rotate180: expected red at (9,9), got %v", c)
	}
}

func TestRotate270(t *testing.T) {
	src := New(20, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := Rotate270(src)
	// 20x10 -> 10x20
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 20 {
		t.Fatalf("Rotate270: got %dx%d, want 10x20", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	// src(0,0) -> dst(0,19) for 270° clockwise
	c := dst.NRGBAAt(0, 19)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("Rotate270: expected red at (0,19), got %v", c)
	}
}

func TestTranspose(t *testing.T) {
	src := New(20, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := Transpose(src)
	// 20x10 -> 10x20
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 20 {
		t.Fatalf("Transpose: got %dx%d, want 10x20", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	// src(0,0) -> dst(0,0)
	c := dst.NRGBAAt(0, 0)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("Transpose: expected red at (0,0), got %v", c)
	}
}

func TestTransverse(t *testing.T) {
	src := New(20, 10, color.Black)
	src.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	dst := Transverse(src)
	// 20x10 -> 10x20
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 20 {
		t.Fatalf("Transverse: got %dx%d, want 10x20", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
	// src(0,0) -> dst(9,19)
	c := dst.NRGBAAt(9, 19)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("Transverse: expected red at (9,19), got %v", c)
	}
}

func TestRotate(t *testing.T) {
	src := New(100, 100, color.NRGBA{R: 255, A: 255})
	dst := Rotate(src, 45, color.Black)
	// Rotated 45° should produce a larger canvas.
	if dst.Bounds().Dx() <= 100 || dst.Bounds().Dy() <= 100 {
		t.Fatalf("Rotate 45: expected larger canvas, got %dx%d",
			dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestRotate_0(t *testing.T) {
	src := New(10, 10, color.White)
	dst := Rotate(src, 0, color.Black)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Fatalf("Rotate 0: expected 10x10, got %dx%d",
			dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestRotate_360(t *testing.T) {
	src := New(10, 10, color.White)
	dst := Rotate(src, 360, color.Black)
	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Fatalf("Rotate 360: expected 10x10, got %dx%d",
			dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}
