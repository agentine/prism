package prism

import (
	"image"
	"image/color"
	"testing"
)

func TestPaste(t *testing.T) {
	bg := New(100, 100, color.Black)
	fg := New(20, 20, color.NRGBA{R: 255, A: 255})
	dst := Paste(bg, fg, image.Pt(10, 10))

	// Pasted area should be red.
	c := dst.NRGBAAt(15, 15)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("expected red at (15,15), got %v", c)
	}

	// Outside pasted area should be black.
	c2 := dst.NRGBAAt(0, 0)
	if c2.R != 0 {
		t.Fatalf("expected black at (0,0), got %v", c2)
	}
}

func TestPaste_OutOfBounds(t *testing.T) {
	bg := New(100, 100, color.Black)
	fg := New(200, 200, color.White)
	// Paste at negative offset — should not panic (fix for #163).
	dst := Paste(bg, fg, image.Pt(-50, -50))
	if dst.Bounds().Dx() != 100 || dst.Bounds().Dy() != 100 {
		t.Fatalf("unexpected size: %dx%d", dst.Bounds().Dx(), dst.Bounds().Dy())
	}
}

func TestPasteCenter(t *testing.T) {
	bg := New(100, 100, color.Black)
	fg := New(20, 20, color.NRGBA{R: 255, A: 255})
	dst := PasteCenter(bg, fg)

	// Center of foreground is at center of background.
	c := dst.NRGBAAt(50, 50)
	if c.R != 255 {
		t.Fatalf("expected red at center, got %v", c)
	}
}

func TestOverlay(t *testing.T) {
	bg := New(100, 100, color.NRGBA{R: 0, G: 0, B: 255, A: 255}) // blue
	fg := New(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255}) // red

	// Full opacity overlay should replace background.
	dst := Overlay(bg, fg, image.Pt(0, 0), 1.0)
	c := dst.NRGBAAt(50, 50)
	if c.R != 255 || c.B != 0 {
		t.Fatalf("expected red at full opacity, got %v", c)
	}

	// Zero opacity should preserve background.
	dst2 := Overlay(bg, fg, image.Pt(0, 0), 0.0)
	c2 := dst2.NRGBAAt(50, 50)
	if c2.B != 255 || c2.R != 0 {
		t.Fatalf("expected blue at zero opacity, got %v", c2)
	}
}

func TestOverlay_HalfOpacity(t *testing.T) {
	bg := New(10, 10, color.NRGBA{R: 0, G: 0, B: 200, A: 255})
	fg := New(10, 10, color.NRGBA{R: 200, G: 0, B: 0, A: 255})
	dst := Overlay(bg, fg, image.Pt(0, 0), 0.5)
	c := dst.NRGBAAt(0, 0)
	// Should be a blend of red and blue.
	if c.R == 0 || c.B == 0 {
		t.Fatalf("expected blend, got %v", c)
	}
}

func TestOverlayCenter(t *testing.T) {
	bg := New(100, 100, color.Black)
	fg := New(20, 20, color.NRGBA{R: 255, A: 128})
	dst := OverlayCenter(bg, fg, 1.0)
	if dst.Bounds().Dx() != 100 {
		t.Fatal("unexpected size")
	}
}
