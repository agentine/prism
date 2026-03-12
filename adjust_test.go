package prism

import (
	"image/color"
	"testing"
)

func TestAdjustBrightness(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255})

	brighter := AdjustBrightness(src, 50)
	c := brighter.NRGBAAt(0, 0)
	if c.R <= 128 {
		t.Fatalf("expected brighter, got R=%d", c.R)
	}

	darker := AdjustBrightness(src, -50)
	c2 := darker.NRGBAAt(0, 0)
	if c2.R >= 128 {
		t.Fatalf("expected darker, got R=%d", c2.R)
	}
}

func TestAdjustContrast(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
	dst := AdjustContrast(src, 50)
	c := dst.NRGBAAt(0, 0)
	// Higher contrast should push values further from midpoint (128).
	if c.R <= 200 {
		t.Fatalf("expected higher contrast, got R=%d", c.R)
	}
}

func TestAdjustGamma(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255})

	// gamma > 1 brightens (raises midtones)
	dst := AdjustGamma(src, 2.0)
	c := dst.NRGBAAt(0, 0)
	if c.R <= 128 {
		t.Fatalf("gamma 2.0 should brighten, got R=%d", c.R)
	}

	// gamma < 1 darkens
	dst2 := AdjustGamma(src, 0.5)
	c2 := dst2.NRGBAAt(0, 0)
	if c2.R >= 128 {
		t.Fatalf("gamma 0.5 should darken, got R=%d", c2.R)
	}
}

func TestAdjustSigmoid(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	dst := AdjustSigmoid(src, 0.5, 5.0)
	// Should not panic and should produce a valid image.
	if dst.Bounds().Dx() != 10 {
		t.Fatal("unexpected size")
	}
}

func TestAdjustSaturation(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	dst := AdjustSaturation(src, -100) // full desaturation
	c := dst.NRGBAAt(0, 0)
	// Fully desaturated should have R≈G≈B.
	diff := absInt(int(c.R)-int(c.G)) + absInt(int(c.G)-int(c.B))
	if diff > 2 {
		t.Fatalf("expected near-gray at -100%% saturation, got %v", c)
	}
}

func TestAdjustHue(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	dst := AdjustHue(src, 120) // red -> green-ish
	c := dst.NRGBAAt(0, 0)
	if c.G <= c.R {
		t.Fatalf("expected hue shift toward green, got %v", c)
	}
}

func TestAdjustFunc(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	dst := AdjustFunc(src, func(c color.NRGBA) color.NRGBA {
		return color.NRGBA{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B, A: c.A}
	})
	c := dst.NRGBAAt(0, 0)
	if c.R != 155 {
		t.Fatalf("expected R=155, got R=%d", c.R)
	}
}

func TestGrayscale(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	dst := Grayscale(src)
	c := dst.NRGBAAt(0, 0)
	if c.R != c.G || c.G != c.B {
		t.Fatalf("expected R=G=B, got %v", c)
	}
}

func TestInvert(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	dst := Invert(src)
	c := dst.NRGBAAt(0, 0)
	if c.R != 55 || c.G != 155 || c.B != 205 {
		t.Fatalf("expected {55,155,205}, got %v", c)
	}
	// Alpha should be unchanged.
	if c.A != 255 {
		t.Fatalf("alpha changed: got %d", c.A)
	}
}

func TestBlur(t *testing.T) {
	// Create image with a single white pixel surrounded by black.
	src := New(21, 21, color.Black)
	src.SetNRGBA(10, 10, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	dst := Blur(src, 2.0)
	// Center should be dimmer after blur.
	c := dst.NRGBAAt(10, 10)
	if c.R >= 255 {
		t.Fatalf("expected blur to reduce center, got R=%d", c.R)
	}
	// A nearby pixel should now have some value.
	c2 := dst.NRGBAAt(11, 10)
	if c2.R == 0 {
		t.Fatal("expected blur to spread to neighbor")
	}
}

func TestSharpen(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	dst := Sharpen(src, 1.0)
	// Uniform image: sharpen should not change it much.
	c := dst.NRGBAAt(5, 5)
	if absInt(int(c.R)-128) > 2 {
		t.Fatalf("expected near-128, got R=%d", c.R)
	}
}

func TestConvolve3x3(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	// Identity-ish kernel (center = 1, rest = 0).
	kernel := [9]float64{0, 0, 0, 0, 1, 0, 0, 0, 0}
	dst := Convolve3x3(src, kernel, nil)
	c := dst.NRGBAAt(5, 5)
	if c.R != 100 || c.G != 100 || c.B != 100 {
		t.Fatalf("identity kernel should preserve color, got %v", c)
	}
}

func TestHistogram(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 128, B: 128, A: 255})
	hist := Histogram(src)
	// All pixels are the same gray, so one bin should have all the weight.
	var total float64
	for _, v := range hist {
		total += v
	}
	if total < 0.99 || total > 1.01 {
		t.Fatalf("histogram should sum to 1.0, got %f", total)
	}
}
