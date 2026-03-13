# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-13

Initial release of **prism** — an image processing library for Go that replaces [disintegration/imaging](https://github.com/disintegration/imaging).

### Added

- **I/O** (`io.go`) — `Open(path)`, `Save(img, path, opts...)`, `Decode(r)`, `Encode(w, img, format, opts...)` supporting JPEG, PNG, GIF, BMP, TIFF, and WebP (decode only). Configurable `MaxImageSize` decompression bomb limit.
- **Resize** (`resize.go`) — `Resize(img, w, h, filter)`, `Fit(img, w, h, filter)`, `Fill(img, w, h, anchor, filter)`, `Thumbnail(img, size, filter)`. Filters: `NearestNeighbor`, `Box`, `Linear`, `Hermite`, `MitchellNetravali`, `CatmullRom`, `BSpline`, `Gaussian`, `Bartlett`, `Lanczos`, `Hann`, `Hamming`, `Blackman`, `Welch`, `Cosine`.
- **Crop & Transform** (`crop.go`, `transform.go`) — `Crop`, `CropAnchor`, `CropCenter`; `Rotate90`, `Rotate180`, `Rotate270`, `FlipH`, `FlipV`, `Rotate(degrees, bg, filter)`, `Transpose`, `Transverse`.
- **Adjustments** (`adjust.go`) — `AdjustBrightness`, `AdjustContrast`, `AdjustGamma`, `AdjustSaturation`, `AdjustHue`, `Grayscale`, `Invert`, `Sigmoid`.
- **Effects** (`effects.go`) — `Blur` (Gaussian), `Sharpen`, `Emboss`, `EdgeDetect`, `UnsharpMask`.
- **Convolution** (`convolution.go`) — `Convolve3x3`, `Convolve5x5` for custom kernel operations.
- **Composition** (`compose.go`) — `Paste(dst, src, pt)`, `PasteCenter`, `Overlay(dst, src, pt, opacity)`, `OverlayCenter`.
- **Histogram** (`histogram.go`) — `Histogram` returning per-channel RGBA frequency counts.
- **EXIF orientation** (`exif.go`) — `AutoOrientation(img, exifData)` applies EXIF rotation/flip tags (orientations 1–8) to produce correctly-oriented images.
- **Helpers** (`helpers.go`) — `New`, `Clone`, `NewRGBA`, pixel sampling helpers.
- **`golang.org/x/image`** — sole dependency, used for BMP/TIFF codec support and extended filter math.
- **Benchmarks** — resize, blur, and composite operation benchmarks.
- **imaging-compatible API** — drop-in replacement for `github.com/disintegration/imaging`.

### Fixed

- `AdjustGamma` direction inversion — was using `pow(x, 1/gamma)` (inverts brightness); corrected to `pow(x, gamma)` matching imaging v1.6.2 semantics (gamma > 1 darkens, gamma < 1 brightens).
- EXIF orientation 6/8 swap — orientation 6 (90° CW) was applying 90° CCW and vice versa.
- Resize filter parameter not threaded through horizontal/vertical passes — filter was being ignored for multi-pass resize operations.
- Decompression bomb check now occurs before full decode to prevent OOM on crafted images.
- WebP decode path: returns explicit `ErrUnsupportedFormat` for WebP encode (encode not supported by stdlib).
