# prism

**Drop-in replacement for `disintegration/imaging` — with CVE-2023-36308 fixed.**

`prism` is a pure-Go image processing library that is 100% API-compatible with
[`disintegration/imaging`](https://github.com/disintegration/imaging) v1.6.2.
Switch with a single import path change and get a maintained library with critical
security fixes, panic bug fixes, and new capabilities including WebP decode and
`AdjustHue`.

## Why prism?

`disintegration/imaging` has 5,700+ stars and nearly 3,000 direct importers, but
has been completely inactive since December 2020. The maintainer has not responded
to issues or pull requests in over five years. In the meantime:

- **CVE-2023-36308** — an index-out-of-bounds in the internal scanner that can be
  triggered by a maliciously crafted image — remains unpatched.
- Multiple confirmed panics (`Rotate90` #159, `Paste` #163, specific images #165)
  have never been fixed.
- `AdjustHue` was merged to master in December 2020 but was never released.
- There is no actively maintained, API-compatible replacement.

`prism` fixes all of these issues while keeping every function signature, type
name, constant value, and variable name identical to `imaging` v1.6.2. Your
existing code compiles without any changes other than the import path.

| Property | disintegration/imaging | prism |
|---|---|---|
| CVE-2023-36308 | Unpatched | Fixed |
| Rotate90 panic (#159) | Unpatched | Fixed |
| Paste panic (#163) | Unpatched | Fixed |
| AdjustHue | Unreleased | Included |
| WebP decode | No | Yes |
| MaxImageSize limit | No | Yes |
| Active maintenance | No | Yes |
| Go version | 1.13 | 1.25+ |
| Dependencies | zero | `golang.org/x/image` |

## Installation

```sh
go get github.com/agentine/prism
```

The only dependency is `golang.org/x/image`, used for BMP, TIFF, and WebP codec
support.

## Quick start

```go
package main

import (
    "image/color"
    "log"

    "github.com/agentine/prism"
)

func main() {
    // Open and decode any supported format (JPEG, PNG, GIF, BMP, TIFF, WebP).
    img, err := prism.Open("photo.jpg", prism.AutoOrientation(true))
    if err != nil {
        log.Fatal(err)
    }

    // Resize to fit within 800x600, preserving aspect ratio.
    resized := prism.Fit(img, 800, 600, prism.Lanczos)

    // Adjust and save.
    adjusted := prism.AdjustBrightness(resized, 10)
    if err := prism.Save(adjusted, "out.jpg", prism.JPEGQuality(90)); err != nil {
        log.Fatal(err)
    }

    // Or compose images.
    background := prism.New(1200, 800, color.White)
    result := prism.PasteCenter(background, resized)
    prism.Save(result, "composed.png")
}
```

## I/O

### Open and Save

```go
// Open reads and decodes an image from a file path.
// Format is detected from the file extension.
img, err := prism.Open("photo.jpg")
img, err  = prism.Open("photo.jpg", prism.AutoOrientation(true))
img, err  = prism.Open("photo.jpg", prism.MaxImageSize(50_000_000)) // 50 MP limit

// Save encodes an image to a file. Format is detected from the extension.
err = prism.Save(img, "out.png")
err = prism.Save(img, "out.jpg", prism.JPEGQuality(85))
err = prism.Save(img, "out.png", prism.PNGCompressionLevel(png.BestCompression))
err = prism.Save(img, "out.gif", prism.GIFNumColors(128))
```

### Decode and Encode

```go
// Decode reads from any io.Reader.
img, err := prism.Decode(r)
img, err  = prism.Decode(r, prism.AutoOrientation(true), prism.MaxImageSize(10_000_000))

// Encode writes to any io.Writer.
err = prism.Encode(w, img, prism.JPEG)
err = prism.Encode(w, img, prism.PNG)
err = prism.Encode(w, img, prism.JPEG, prism.JPEGQuality(95))
```

### Supported formats

| Format | Decode | Encode |
|---|---|---|
| JPEG | Yes | Yes |
| PNG | Yes | Yes |
| GIF | Yes | Yes |
| BMP | Yes | Yes |
| TIFF | Yes | Yes |
| WebP | Yes | No* |

*WebP encoding requires cgo and is not available in a pure-Go build.
`Encode(w, img, prism.WEBP)` returns `ErrUnsupportedFormat`.

### Encode options

```go
prism.JPEGQuality(quality int)               // 1–100, default 95
prism.PNGCompressionLevel(png.CompressionLevel)
prism.GIFNumColors(numColors int)            // 1–256, default 256
prism.GIFQuantizer(draw.Quantizer)
prism.GIFDrawer(draw.Drawer)
```

### Format helpers

```go
format, err := prism.FormatFromExtension(".jpg")   // prism.JPEG
format, err  = prism.FormatFromFilename("photo.png") // prism.PNG
```

### Errors

```go
prism.ErrUnsupportedFormat // unrecognised extension or format
prism.ErrImageTooLarge     // image exceeds MaxImageSize pixel limit
```

## Resize and resample

### Functions

```go
// Resize to exact dimensions. Pass 0 for one dimension to preserve aspect ratio.
// Both 0 returns a clone of the original.
dst := prism.Resize(img, 800, 600, prism.Lanczos)
dst  = prism.Resize(img, 800, 0, prism.Lanczos)   // height calculated from aspect ratio

// Fit scales the image to fit within the given bounds, never enlarging.
dst = prism.Fit(img, 800, 600, prism.Lanczos)

// Fill scales and crops to exactly fill the given bounds.
// Anchor controls which region is kept after cropping.
dst = prism.Fill(img, 400, 400, prism.Center, prism.Lanczos)
dst = prism.Fill(img, 400, 400, prism.Top, prism.Lanczos)

// Thumbnail is Fill with Center anchor — the standard thumbnail operation.
dst = prism.Thumbnail(img, 256, 256, prism.Lanczos)
```

### Resample filters

Choose quality vs. speed:

| Filter | Quality | Speed | Notes |
|---|---|---|---|
| `prism.NearestNeighbor` | Lowest | Fastest | Pixel art, no interpolation |
| `prism.Box` | Low | Fast | Simple averaging |
| `prism.Linear` | Medium | Fast | Bilinear interpolation |
| `prism.Hermite` | Medium | Fast | Smooth bilinear variant |
| `prism.MitchellNetravali` | High | Medium | Balanced quality/ringing |
| `prism.CatmullRom` | High | Medium | Sharp with mild ringing |
| `prism.BSpline` | Medium | Medium | Very smooth, softer result |
| `prism.Gaussian` | Medium | Medium | Smooth, slightly blurry |
| `prism.Bartlett` | Medium | Medium | Triangle window |
| `prism.Lanczos` | Highest | Slower | Best for downscaling |
| `prism.Hann` | High | Slower | Windowed sinc |
| `prism.Hamming` | High | Slower | Windowed sinc |
| `prism.Blackman` | High | Slower | Windowed sinc |
| `prism.Welch` | High | Slower | Parabolic window |
| `prism.Cosine` | High | Slower | Cosine window |

For most photographic downscaling, use `prism.Lanczos`. For upscaling or speed-
sensitive paths, use `prism.CatmullRom` or `prism.Linear`.

## Crop

```go
// Crop to an arbitrary rectangle.
dst := prism.Crop(img, image.Rect(10, 10, 200, 150))

// Crop to a centered rectangle.
dst = prism.CropCenter(img, 400, 300)

// Crop anchored to any corner or edge.
dst = prism.CropAnchor(img, 400, 300, prism.TopLeft)
dst = prism.CropAnchor(img, 400, 300, prism.BottomRight)
```

### Anchor constants

```go
prism.Center      // default
prism.TopLeft
prism.Top
prism.TopRight
prism.Left
prism.Right
prism.BottomLeft
prism.Bottom
prism.BottomRight
```

## Rotate and flip

```go
// Fixed rotations (lossless, no interpolation).
dst := prism.Rotate90(img)    // 90° clockwise
dst  = prism.Rotate180(img)
dst  = prism.Rotate270(img)   // 90° counter-clockwise

// Mirror.
dst = prism.FlipH(img) // horizontal flip
dst = prism.FlipV(img) // vertical flip

// Diagonal transpose operations.
dst = prism.Transpose(img)  // flip along top-left to bottom-right diagonal
dst = prism.Transverse(img) // flip along top-right to bottom-left diagonal

// Arbitrary angle (bilinear interpolation, expands canvas to fit).
dst = prism.Rotate(img, 45.0, color.White)  // degrees counter-clockwise, background fill
```

## Blur and sharpen

```go
// Gaussian blur. sigma is the standard deviation of the kernel.
// Larger sigma = stronger blur.
dst := prism.Blur(img, 3.0)

// Unsharp mask sharpening. sigma controls the blur radius.
// Larger sigma = stronger sharpening effect.
dst = prism.Sharpen(img, 1.5)
```

## Color adjustments

All percentage-based functions accept negative values for the inverse effect.

```go
// Brightness: percentage in [-100, 100]. Positive = brighter.
dst := prism.AdjustBrightness(img, 20)
dst  = prism.AdjustBrightness(img, -15)

// Contrast: percentage in [-100, 100]. Positive = more contrast.
dst = prism.AdjustContrast(img, 30)

// Gamma correction. gamma > 1 darkens, gamma < 1 brightens.
dst = prism.AdjustGamma(img, 1.8)

// Saturation: percentage in [-100, 500]. -100 = grayscale, 0 = unchanged.
dst = prism.AdjustSaturation(img, 50)
dst = prism.AdjustSaturation(img, -100) // grayscale via saturation

// Hue rotation in degrees. Wraps around at 360.
dst = prism.AdjustHue(img, 90)
dst = prism.AdjustHue(img, -45)

// Sigmoid contrast. midpoint in [0, 1], factor controls steepness.
dst = prism.AdjustSigmoid(img, 0.5, 5.0)

// Custom per-pixel transformation.
dst = prism.AdjustFunc(img, func(c color.NRGBA) color.NRGBA {
    c.R = 255 - c.R
    return c
})
```

## Grayscale and invert

```go
// Convert to grayscale using ITU-R BT.601 luminance coefficients.
dst := prism.Grayscale(img)

// Invert all color channels.
dst = prism.Invert(img)
```

## Convolution

Apply custom kernels for edge detection, emboss, or any linear filter:

```go
// 3x3 kernel (row-major order).
kernel3 := [9]float64{
    -1, -1, -1,
    -1,  8, -1,
    -1, -1, -1,
}
dst := prism.Convolve3x3(img, kernel3, &prism.ConvolveOptions{Normalize: false})

// 5x5 kernel.
var kernel5 [25]float64
// ... fill kernel5 ...
dst = prism.Convolve5x5(img, kernel5, nil) // nil options = defaults

// ConvolveOptions fields:
// Normalize bool — divide each weight by the kernel sum before applying
// Abs       bool — take absolute value of each output channel
// Bias      int  — constant added to each channel after convolution
```

## Composition

```go
// Paste src over background at pixel position (x, y). No alpha blending.
dst := prism.Paste(background, src, image.Pt(50, 100))

// Paste centered.
dst = prism.PasteCenter(background, src)

// Alpha-blended overlay. opacity in [0.0, 1.0].
dst = prism.Overlay(background, src, image.Pt(50, 100), 0.75)

// Overlay centered.
dst = prism.OverlayCenter(background, src, 0.5)
```

`Overlay` uses Porter-Duff "over" compositing with per-pixel alpha. `Paste`
performs a straight copy with no alpha blending.

## Image creation and cloning

```go
// New creates a blank image filled with a solid color.
blank := prism.New(1920, 1080, color.White)
transparent := prism.New(800, 600, color.Transparent)

// Clone returns a copy of any image.Image as *image.NRGBA.
copy := prism.Clone(img)
```

## Histogram

```go
// Returns a normalized luminance histogram [256]float64.
// Each entry is the fraction of pixels at that luminance level.
hist := prism.Histogram(img)
fmt.Printf("%.3f%% of pixels are at luminance 128\n", hist[128]*100)
```

## EXIF auto-orientation

JPEG files shot on phones are often stored rotated with an EXIF orientation tag
that tells viewers how to display them. The original `imaging` library had a
known bug where orientations 6 and 8 (90° CW and 90° CCW) were swapped.
`prism` reads the EXIF orientation from the APP1 segment and applies the correct
transformation.

```go
// AutoOrientation(true) rotates/flips the image to match its EXIF tag.
img, err := prism.Open("photo.jpg", prism.AutoOrientation(true))

// Or with Decode:
img, err = prism.Decode(r, prism.AutoOrientation(true))
```

Orientation values 1–8 are all handled correctly. Orientation 1 (normal) is a
no-op. The EXIF orientation 6/8 swap present in `disintegration/imaging` is fixed.

## Security

### CVE-2023-36308

**CVE-2023-36308** is an index-out-of-bounds vulnerability in the `scan` function
of `disintegration/imaging`. A maliciously crafted image can cause the scanner to
read outside the allocated pixel buffer, resulting in a panic or potential memory
corruption.

The root cause is that the scanner in `imaging` does not clamp the requested
rectangle to the actual image bounds before accessing `Pix`. Any image whose
reported dimensions differ from its actual pixel buffer size — including truncated
or malformed images — can trigger the panic.

`prism` fixes this at the scanner level in `scanner.go`. Every call to `scan`
clamps the requested rectangle to the image bounds before any pixel access:

```
x1 = max(x1, bounds.Min.X)
y1 = max(y1, bounds.Min.Y)
x2 = min(x2, bounds.Max.X)
y2 = min(y2, bounds.Max.Y)
```

All image processing functions — resize, crop, rotate, blur, adjust, compose —
route pixel access through this scanner, so the entire operation pipeline is
protected by a single bounds check.

### Decompression bomb protection

A decompression bomb (zip bomb for images) is a small compressed file that
expands to a very large pixel buffer. For example, a 1 MB PNG can decompress to
hundreds of megabytes of pixel data. Without a limit, processing user-supplied
images can exhaust available memory.

`prism` provides `MaxImageSize` to reject images before allocating their pixel
data:

```go
// Reject any image whose width * height exceeds 50 million pixels (~180 MB NRGBA).
img, err := prism.Open("untrusted.png", prism.MaxImageSize(50_000_000))
if err == prism.ErrImageTooLarge {
    // reject the request
}
```

When `MaxImageSize` is set, `Decode` calls `image.DecodeConfig` first to read
only the image dimensions (no pixel allocation), then rejects the image if it
exceeds the limit, and only proceeds to full decode if the check passes.

The internal `New` function also enforces a hard cap of 100 megapixels on all
internally allocated images regardless of `MaxImageSize`.

## Migration from disintegration/imaging

For the vast majority of users, migration is a single import path change followed
by renaming the package alias:

```go
// Before
import "github.com/disintegration/imaging"
result := imaging.Resize(img, 800, 600, imaging.Lanczos)

// After
import "github.com/agentine/prism"
result := prism.Resize(img, 800, 600, prism.Lanczos)
```

To update an entire module automatically:

```sh
# Replace the import path in all .go files
find . -name '*.go' | xargs sed -i '' \
  's|github.com/disintegration/imaging|github.com/agentine/prism|g'

# Update all package references (if you used "imaging" as the identifier)
find . -name '*.go' | xargs sed -i '' \
  's/\bimaging\./prism./g'

# Tidy
go get github.com/agentine/prism
go mod tidy
```

### Behavior differences

The following differences from `disintegration/imaging` v1.6.2 are intentional
corrections of bugs in the original:

| Behavior | imaging v1.6.2 | prism |
|---|---|---|
| EXIF orientation 6 | Rotates 270° CCW | Rotates 90° CW (correct) |
| EXIF orientation 8 | Rotates 90° CW | Rotates 270° CCW (correct) |
| `AdjustGamma(img, gamma)` | `pow(x, 1/gamma)` | `pow(x, gamma)` (matches docs) |
| Malformed image panic | Panics (CVE-2023-36308) | Returns error or zero image |
| `Paste` out-of-bounds | Panics | Clips to background bounds |

### New in prism

The following are available in `prism` but not in `disintegration/imaging` v1.6.2:

- `AdjustHue(img, shift float64)` — hue rotation in degrees (was in imaging master but never released)
- `MaxImageSize(pixels int)` — `DecodeOption` for decompression bomb protection
- `ErrImageTooLarge` — error returned when `MaxImageSize` is exceeded
- WebP decode support (via `golang.org/x/image/webp`)

## API reference

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
func MaxImageSize(pixels int) DecodeOption
func JPEGQuality(quality int) EncodeOption
func PNGCompressionLevel(level png.CompressionLevel) EncodeOption
func GIFNumColors(numColors int) EncodeOption
func GIFQuantizer(quantizer draw.Quantizer) EncodeOption
func GIFDrawer(drawer draw.Drawer) EncodeOption
```

### Functions — creation and cloning

```go
func New(width, height int, fillColor color.Color) *image.NRGBA
func Clone(img image.Image) *image.NRGBA
```

### Functions — resize and scale

```go
func Resize(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
func Fit(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
func Fill(img image.Image, width, height int, anchor Anchor, filter ResampleFilter) *image.NRGBA
func Thumbnail(img image.Image, width, height int, filter ResampleFilter) *image.NRGBA
```

### Functions — crop

```go
func Crop(img image.Image, rect image.Rectangle) *image.NRGBA
func CropCenter(img image.Image, width, height int) *image.NRGBA
func CropAnchor(img image.Image, width, height int, anchor Anchor) *image.NRGBA
```

### Functions — transform

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

### Functions — adjustments

```go
func AdjustBrightness(img image.Image, percentage float64) *image.NRGBA
func AdjustContrast(img image.Image, percentage float64) *image.NRGBA
func AdjustGamma(img image.Image, gamma float64) *image.NRGBA
func AdjustSaturation(img image.Image, percentage float64) *image.NRGBA
func AdjustSigmoid(img image.Image, midpoint, factor float64) *image.NRGBA
func AdjustHue(img image.Image, shift float64) *image.NRGBA
func AdjustFunc(img image.Image, fn func(c color.NRGBA) color.NRGBA) *image.NRGBA
func Grayscale(img image.Image) *image.NRGBA
func Invert(img image.Image) *image.NRGBA
```

### Functions — blur and convolution

```go
func Blur(img image.Image, sigma float64) *image.NRGBA
func Sharpen(img image.Image, sigma float64) *image.NRGBA
func Convolve3x3(img image.Image, kernel [9]float64, options *ConvolveOptions) *image.NRGBA
func Convolve5x5(img image.Image, kernel [25]float64, options *ConvolveOptions) *image.NRGBA
```

### Functions — composition

```go
func Paste(background, img image.Image, pos image.Point) *image.NRGBA
func PasteCenter(background, img image.Image) *image.NRGBA
func Overlay(background, img image.Image, pos image.Point, opacity float64) *image.NRGBA
func OverlayCenter(background, img image.Image, opacity float64) *image.NRGBA
```

### Functions — analysis

```go
func Histogram(img image.Image) [256]float64
```

## License

MIT. See [LICENSE](LICENSE).
