# Prism — Design & Implementation Plan

**Package:** `github.com/agentine/prism`
**Registry:** Go module proxy (pkg.go.dev)
**Replaces:** `disintegration/imaging` (5.7K stars, 2,899 importers, last commit Dec 2020, last release Nov 2019)

## Problem

`disintegration/imaging` is the most widely used pure-Go image processing library. It provides resize, crop, rotate, blur, sharpen, color adjustments, and image composition — all without C dependencies. Despite 2,899 direct importers and 5.7K stars, the project has been completely inactive for over 5 years:

- **Single maintainer** (Grigory Dryapak) has not responded to issues since 2020
- **CVE-2023-36308** — index out of bounds in scan function with malicious images (UNPATCHED)
- **Multiple panic bugs** — Rotate90 panic (#159), Paste panic (#163), crashes with specific images (#165)
- **Zero funding**, no GitHub Sponsors or OpenCollective
- **Community asking "is this maintained?"** — Issue #169 (Feb 2024), Issue #168 asking for v1.6.3 (Jan 2024), no maintainer response
- **Unreleased code on master** — AdjustHue function merged Dec 2020 but never released
- **No API-compatible replacement** — bild (4.1K stars) has a completely different API; gift is by the same inactive author; nfnt/resize is archived and only does resizing

## Goals

1. **100% API-compatible** with `disintegration/imaging` v1.6.2 — same function signatures, types, constants, variables
2. **Single import path change** for migration: `imaging` → `prism`
3. **Fix all known bugs** — CVE-2023-36308, panics, edge cases
4. **Security by default** — bounds checking, safe malicious image handling, configurable memory limits
5. **Performance improvements** — parallel processing, reduced allocations, SIMD-friendly algorithms
6. **New capabilities** — WebP support, AdjustHue, configurable memory limits

## API Surface (100% Compatible)

### Constants

```go
// Anchor points for cropping and filling
const (
    Center Anchor = iota
    TopLeft
    Top
    TopRight
    Left
    Right
    BottomLeft
    Bottom
    BottomRight
)

// Image formats
const (
    JPEG Format = iota
    PNG
    GIF
    TIFF
    BMP
)
```

### Variables

```go
var ErrUnsupportedFormat = errors.New("imaging: unsupported image format")

// Resample filters (same names, same behavior)
var (
    NearestNeighbor    ResampleFilter
    Box                ResampleFilter
    Linear             ResampleFilter
    Hermite            ResampleFilter
    MitchellNetravali  ResampleFilter
    CatmullRom         ResampleFilter
    BSpline            ResampleFilter
    Gaussian           ResampleFilter
    Bartlett           ResampleFilter
    Lanczos            ResampleFilter
    Hann               ResampleFilter
    Hamming            ResampleFilter
    Blackman           ResampleFilter
    Welch              ResampleFilter
    Cosine             ResampleFilter
)
```

### Types

```go
type Anchor int
type Format int
type ResampleFilter struct {
    Support float64
    Kernel  func(float64) float64
}
type ConvolveOptions struct {
    Normalize bool
    Abs       bool
    Bias      int
}
type DecodeOption func(*decodeConfig)
type EncodeOption func(*encodeConfig)
```

### Functions — I/O

```go
func Open(filename string, opts ...DecodeOption) (image.Image, error)
func Decode(r io.Reader, opts ...DecodeOption) (image.Image, error)
func Save(img image.Image, filename string, opts ...EncodeOption) error
func Encode(w io.Writer, img image.Image, format Format, opts ...EncodeOption) error
func FormatFromExtension(ext string) (Format, error)
func FormatFromFilename(filename string) (Format, error)
func AutoOrientation(enabled bool) DecodeOption
func JPEGQuality(quality int) EncodeOption
func PNGCompressionLevel(level png.CompressionLevel) EncodeOption
func GIFNumColors(numColors int) EncodeOption
func GIFQuantizer(quantizer draw.Quantizer) EncodeOption
func GIFDrawer(drawer draw.Drawer) EncodeOption
```

### Functions — Creation & Cloning

```go
func New(width, height int, fillColor color.Color) *image.NRGBA
func Clone(img image.Image) *image.NRGBA
```

### Functions — Resize & Scale

```go
func Resize(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
func Fit(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
func Fill(img image.Image, width, height int, anchor Anchor, filter ResampleFilter) *image.NRGBA
func Thumbnail(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
```

### Functions — Crop

```go
func Crop(img image.Image, rect image.Rectangle) *image.NRGBA
func CropCenter(img image.Image, width, height int) *image.NRGBA
func CropAnchor(img image.Image, width, height int, anchor Anchor) *image.NRGBA
```

### Functions — Transform

```go
func Rotate(img image.Image, angle float64, bgColor color.Color) *image.NRGBA
func Rotate90(img image.Image) *image.NRGBA
func Rotate180(img image.Image) *image.NRGBA
func Rotate270(img image.Image) *image.NRGBA
func FlipH(img image.Image) *image.NRGBA
func FlipV(img image.Image) *image.NRGBA
func Transpose(img image.Image) *image.NRGBA
func Transverse(img image.Image) *image.NRGBA
```

### Functions — Adjustments

```go
func AdjustBrightness(img image.Image, percentage float64) *image.NRGBA
func AdjustContrast(img image.Image, percentage float64) *image.NRGBA
func AdjustGamma(img image.Image, gamma float64) *image.NRGBA
func AdjustSaturation(img image.Image, percentage float64) *image.NRGBA
func AdjustSigmoid(img image.Image, midpoint, factor float64) *image.NRGBA
func AdjustHue(img image.Image, shift float64) *image.NRGBA  // NEW (unreleased in imaging)
func AdjustFunc(img image.Image, fn func(c color.NRGBA) color.NRGBA) *image.NRGBA
func Grayscale(img image.Image) *image.NRGBA
func Invert(img image.Image) *image.NRGBA
```

### Functions — Filters & Effects

```go
func Blur(img image.Image, sigma float64) *image.NRGBA
func Sharpen(img image.Image, sigma float64) *image.NRGBA
func Convolve3x3(img image.Image, kernel [9]float64, options *ConvolveOptions) *image.NRGBA
func Convolve5x5(img image.Image, kernel [25]float64, options *ConvolveOptions) *image.NRGBA
```

### Functions — Composition

```go
func Paste(background, img image.Image, pos image.Point) *image.NRGBA
func PasteCenter(background, img image.Image) *image.NRGBA
func Overlay(background, img image.Image, pos image.Point, opacity float64) *image.NRGBA
func OverlayCenter(background, img image.Image, opacity float64) *image.NRGBA
```

### Functions — Analysis

```go
func Histogram(img image.Image) [256]float64
```

## New Features (Beyond Compatibility)

### WebP Support (New Format)

```go
const (
    JPEG Format = iota
    PNG
    GIF
    TIFF
    BMP
    WEBP  // NEW
)

func WebPQuality(quality int) EncodeOption  // NEW
```

WebP is the most requested missing format. Encode via `x/image/webp` or bundled encoder.

### Configurable Memory Limits

```go
func MaxImageSize(pixels int) DecodeOption  // NEW — reject images exceeding pixel count
```

Prevents memory exhaustion attacks from decompression bombs (related to CVE-2023-36308).

### AdjustHue

```go
func AdjustHue(img image.Image, shift float64) *image.NRGBA
```

Was merged to imaging master in Dec 2020 but never released. We include it from day one.

## Bug Fixes

| Bug | imaging issue | Fix |
|-----|--------------|-----|
| CVE-2023-36308: index out of bounds in scan with malicious images | #179 | Bounds checking before pixel access in scanner |
| Rotate90 panic with certain image dimensions | #159 | Validate dimensions, handle edge cases |
| Paste function panic | #163 | Bounds intersection before copy |
| Index out of bounds with specific images | #165 | Safe pixel access throughout pipeline |
| Rotated image background blending | #164 | Correct alpha blending in rotation fill |
| GIF animation handling | #173 | Proper frame iteration and disposal |

## Architecture

```
prism/
├── prism.go          # Package doc, format types, anchor constants
├── io.go             # Open, Decode, Save, Encode, format detection
├── scanner.go        # Internal pixel scanner (FIXED: bounds checking)
├── resize.go         # Resize, Fit, Fill, Thumbnail + resample filters
├── crop.go           # Crop, CropCenter, CropAnchor
├── transform.go      # Rotate, Rotate90/180/270, Flip, Transpose
├── adjust.go         # Brightness, Contrast, Gamma, Saturation, Sigmoid, Hue, Func
├── effects.go        # Blur, Sharpen, Grayscale, Invert
├── convolution.go    # Convolve3x3, Convolve5x5
├── compose.go        # Paste, PasteCenter, Overlay, OverlayCenter
├── histogram.go      # Histogram
├── helpers.go        # Internal utilities (clamp, parallel, etc.)
├── go.mod
├── go.sum
├── LICENSE           # MIT
└── README.md
```

## Implementation Phases

### Phase 1: Core Engine & I/O

- Package scaffolding: `go.mod`, LICENSE (MIT), basic README
- Internal scanner with bounds-safe pixel access (CVE-2023-36308 fix)
- `New()`, `Clone()` — image creation
- Format type, Anchor type, ResampleFilter type, ConvolveOptions type
- `FormatFromExtension()`, `FormatFromFilename()`
- `Open()`, `Decode()` with `AutoOrientation` option
- `Save()`, `Encode()` with all encode options (JPEGQuality, PNGCompressionLevel, GIF options)
- `MaxImageSize` decode option (new)
- All format support: JPEG, PNG, GIF, TIFF, BMP
- Unit tests for I/O roundtrips, format detection, auto-orientation

### Phase 2: Resize, Crop & Transform

- All 15 resample filters with identical kernel functions
- `Resize()` — width/height with filter, handles 0-value auto-calculate
- `Fit()` — scale to fit within bounds preserving aspect ratio
- `Fill()` — scale and crop to fill bounds using anchor
- `Thumbnail()` — alias for Fill with appropriate defaults
- `Crop()`, `CropCenter()`, `CropAnchor()`
- `Rotate()` — arbitrary angle rotation with background color
- `Rotate90()`, `Rotate180()`, `Rotate270()` — fixed rotations (fix panic #159)
- `FlipH()`, `FlipV()` — horizontal/vertical mirror
- `Transpose()`, `Transverse()` — diagonal flips
- Parallel processing for resize operations (runtime.NumCPU)
- Unit tests + golden image tests comparing output to imaging v1.6.2

### Phase 3: Adjustments & Effects

- `AdjustBrightness()`, `AdjustContrast()` — percentage-based
- `AdjustGamma()` — gamma correction with lookup table
- `AdjustSaturation()` — HSL-based saturation adjustment
- `AdjustSigmoid()` — sigmoid contrast enhancement
- `AdjustHue()` — hue rotation (new, from imaging master)
- `AdjustFunc()` — custom per-pixel color transformation
- `Grayscale()`, `Invert()` — color transformations
- `Blur()` — Gaussian blur via separable kernel
- `Sharpen()` — unsharp mask via blur + blend
- `Convolve3x3()`, `Convolve5x5()` — general convolution with options
- `Histogram()` — luminance histogram
- Unit tests for all adjustment functions

### Phase 4: Composition, Safety & Polish

- `Paste()`, `PasteCenter()` — image composition (fix panic #163)
- `Overlay()`, `OverlayCenter()` — alpha-blended composition
- WebP encode/decode support (new format)
- Fuzz testing with `go test -fuzz` for all decode paths and transform functions
- Compatibility golden tests: process identical inputs through both imaging and prism, compare output pixel-by-pixel
- Benchmarks vs imaging v1.6.2 for all major operations
- Migration guide (import path swap + behavior notes)
- README with examples, benchmarks, migration section
- CI/CD pipeline (GitHub Actions)
- Go module publish via tag

## Technical Decisions

- **Go 1.21+** — modern minimum, enables slices/maps packages, improved generics
- **Zero dependencies** — pure Go, only stdlib + `golang.org/x/image` for extended format support
- **Parallel processing** — resize and blur operations use `runtime.NumCPU()` goroutines
- **Bounds-safe scanner** — all pixel access goes through bounds-checked scanner, eliminating the entire class of CVE-2023-36308 bugs
- **Lookup tables** — precompute gamma/brightness/contrast tables for O(1) per-pixel adjustments
- **sync.Pool buffers** — reduce GC pressure for intermediate image buffers
- **Error string prefix** — keep `"imaging: "` prefix in error messages for backward compatibility
- **MIT license** — same as original

## Migration

For most users, migration is a single import path change:

```go
// Before
import "github.com/disintegration/imaging"
result := imaging.Resize(img, 800, 600, imaging.Lanczos)

// After
import "github.com/agentine/prism"
result := prism.Resize(img, 800, 600, prism.Lanczos)
```

All function signatures, type names, constant values, and variable names are identical. Output images are pixel-identical for the same inputs (verified by golden tests).
