package prism

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestNew(t *testing.T) {
	img := New(100, 50, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 50 {
		t.Fatalf("size: got %dx%d, want 100x50", img.Bounds().Dx(), img.Bounds().Dy())
	}
	c := img.NRGBAAt(0, 0)
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Fatalf("color: got %v, want red", c)
	}
}

func TestNew_Zero(t *testing.T) {
	img := New(0, 0, color.Black)
	if img.Bounds().Dx() != 0 || img.Bounds().Dy() != 0 {
		t.Fatalf("expected empty image for zero size")
	}
}

func TestClone(t *testing.T) {
	src := New(10, 10, color.NRGBA{R: 128, G: 64, B: 32, A: 200})
	dst := Clone(src)

	if dst.Bounds().Dx() != 10 || dst.Bounds().Dy() != 10 {
		t.Fatalf("size mismatch")
	}
	c := dst.NRGBAAt(5, 5)
	if c.R != 128 || c.G != 64 || c.B != 32 || c.A != 200 {
		t.Fatalf("color: got %v, want {128,64,32,200}", c)
	}
}

func TestClone_Nil(t *testing.T) {
	dst := Clone(nil)
	if dst == nil {
		t.Fatal("expected non-nil empty image")
	}
}

func TestFormatFromExtension(t *testing.T) {
	tests := []struct {
		ext    string
		format Format
		ok     bool
	}{
		{".jpg", JPEG, true},
		{".jpeg", JPEG, true},
		{".JPG", JPEG, true},
		{".png", PNG, true},
		{".gif", GIF, true},
		{".tiff", TIFF, true},
		{".tif", TIFF, true},
		{".bmp", BMP, true},
		{".xyz", 0, false},
	}

	for _, tt := range tests {
		f, err := FormatFromExtension(tt.ext)
		if tt.ok {
			if err != nil {
				t.Errorf("ext=%s: unexpected error: %v", tt.ext, err)
			}
			if f != tt.format {
				t.Errorf("ext=%s: got %d, want %d", tt.ext, f, tt.format)
			}
		} else {
			if err == nil {
				t.Errorf("ext=%s: expected error", tt.ext)
			}
		}
	}
}

func TestFormatFromFilename(t *testing.T) {
	f, err := FormatFromFilename("photo.png")
	if err != nil || f != PNG {
		t.Fatalf("got %d/%v, want PNG", f, err)
	}
}

func TestEncodeDecodePNG(t *testing.T) {
	src := New(8, 8, color.NRGBA{R: 100, G: 200, B: 50, A: 255})

	var buf bytes.Buffer
	if err := Encode(&buf, src, PNG); err != nil {
		t.Fatalf("encode: %v", err)
	}

	img, err := Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
		t.Fatalf("size mismatch after roundtrip")
	}
}

func TestEncodeDecodeJPEG(t *testing.T) {
	src := New(8, 8, color.NRGBA{R: 100, G: 200, B: 50, A: 255})

	var buf bytes.Buffer
	if err := Encode(&buf, src, JPEG, JPEGQuality(90)); err != nil {
		t.Fatalf("encode: %v", err)
	}

	img, err := Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
		t.Fatalf("size mismatch after roundtrip")
	}
}

func TestEncodeDecodeGIF(t *testing.T) {
	src := New(8, 8, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	var buf bytes.Buffer
	if err := Encode(&buf, src, GIF, GIFNumColors(16)); err != nil {
		t.Fatalf("encode: %v", err)
	}

	img, err := Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if img.Bounds().Dx() != 8 || img.Bounds().Dy() != 8 {
		t.Fatalf("size mismatch after roundtrip")
	}
}

func TestMaxImageSize(t *testing.T) {
	// Create a small PNG in memory.
	src := New(100, 100, color.White)
	var buf bytes.Buffer
	png.Encode(&buf, src)

	// Should succeed with large limit.
	_, err := Decode(bytes.NewReader(buf.Bytes()), MaxImageSize(100000))
	if err != nil {
		t.Fatalf("expected success with large limit: %v", err)
	}

	// Should fail with small limit.
	_, err = Decode(bytes.NewReader(buf.Bytes()), MaxImageSize(100))
	if err == nil {
		t.Fatal("expected error with small limit")
	}
}

func TestScannerBoundsCheck(t *testing.T) {
	// Create an NRGBA image with known bounds.
	img := image.NewNRGBA(image.Rect(10, 10, 20, 20))
	for i := range img.Pix {
		img.Pix[i] = 128
	}

	s := newScanner(img)

	// Scan within bounds — should not panic.
	dst := make([]byte, 10*10*4)
	s.scan(10, 10, 20, 20, dst)

	// Scan with coordinates exceeding bounds — should not panic.
	dst2 := make([]byte, 20*20*4)
	s.scan(0, 0, 30, 30, dst2) // way outside bounds — must not panic
}

func TestResampleFilterKernels(t *testing.T) {
	// Verify that all filter kernels return 0 outside their support.
	filters := []struct {
		name   string
		filter ResampleFilter
	}{
		{"Box", Box},
		{"Linear", Linear},
		{"Hermite", Hermite},
		{"MitchellNetravali", MitchellNetravali},
		{"CatmullRom", CatmullRom},
		{"BSpline", BSpline},
		{"Gaussian", Gaussian},
		{"Bartlett", Bartlett},
		{"Lanczos", Lanczos},
		{"Hann", Hann},
		{"Hamming", Hamming},
		{"Blackman", Blackman},
		{"Welch", Welch},
		{"Cosine", Cosine},
	}

	for _, f := range filters {
		// Should be non-zero near 0.
		v := f.filter.Kernel(0)
		if v == 0 {
			t.Errorf("%s: kernel(0) = 0, want non-zero", f.name)
		}
		// Should be zero far outside support.
		v = f.filter.Kernel(f.filter.Support + 1)
		if v != 0 {
			t.Errorf("%s: kernel(%f) = %f, want 0", f.name, f.filter.Support+1, v)
		}
	}
}

// Verify decode of various formats by encoding and decoding.
func TestDecodeJPEG(t *testing.T) {
	src := New(4, 4, color.NRGBA{R: 200, G: 100, B: 50, A: 255})
	var buf bytes.Buffer
	jpeg.Encode(&buf, src, nil)
	img, err := Decode(&buf)
	if err != nil {
		t.Fatalf("decode JPEG: %v", err)
	}
	if img.Bounds().Dx() != 4 {
		t.Fatalf("wrong width")
	}
}
