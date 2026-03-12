package prism

import (
	"encoding/binary"
	"image"
	"io"
)

// readOrientation reads the EXIF orientation tag from a JPEG stream.
// Returns 0 if the orientation cannot be determined.
func readOrientation(r io.Reader) int {
	// JPEG starts with SOI marker 0xFF 0xD8.
	var soi [2]byte
	if _, err := io.ReadFull(r, soi[:]); err != nil || soi[0] != 0xFF || soi[1] != 0xD8 {
		return 0
	}

	// Scan markers until we find APP1 (0xFF 0xE1) or a non-APP marker.
	for {
		var marker [2]byte
		if _, err := io.ReadFull(r, marker[:]); err != nil {
			return 0
		}
		if marker[0] != 0xFF {
			return 0
		}

		// Skip padding 0xFF bytes.
		for marker[1] == 0xFF {
			if _, err := io.ReadFull(r, marker[1:]); err != nil {
				return 0
			}
		}

		// SOS or EOI means no more metadata.
		if marker[1] == 0xDA || marker[1] == 0xD9 {
			return 0
		}

		// Read segment length (big-endian, includes the 2 length bytes).
		var lenBuf [2]byte
		if _, err := io.ReadFull(r, lenBuf[:]); err != nil {
			return 0
		}
		segLen := int(binary.BigEndian.Uint16(lenBuf[:])) - 2
		if segLen < 0 {
			return 0
		}

		if marker[1] != 0xE1 {
			// Not APP1 — skip this segment.
			if _, err := io.CopyN(io.Discard, r, int64(segLen)); err != nil {
				return 0
			}
			continue
		}

		// APP1 segment found. Read the entire segment.
		seg := make([]byte, segLen)
		if _, err := io.ReadFull(r, seg); err != nil {
			return 0
		}

		return parseExifOrientation(seg)
	}
}

// parseExifOrientation parses an APP1 segment payload for the orientation tag.
func parseExifOrientation(data []byte) int {
	// Check "Exif\x00\x00" header.
	if len(data) < 14 || string(data[:6]) != "Exif\x00\x00" {
		return 0
	}

	tiff := data[6:]

	// Determine byte order.
	var bo binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 0
	}

	// Verify TIFF magic number 42.
	if bo.Uint16(tiff[2:4]) != 0x002A {
		return 0
	}

	// Offset to IFD0.
	ifdOffset := int(bo.Uint32(tiff[4:8]))
	if ifdOffset < 8 || ifdOffset+2 > len(tiff) {
		return 0
	}

	numEntries := int(bo.Uint16(tiff[ifdOffset : ifdOffset+2]))
	ifdOffset += 2

	for i := 0; i < numEntries; i++ {
		entryOff := ifdOffset + i*12
		if entryOff+12 > len(tiff) {
			return 0
		}

		tag := bo.Uint16(tiff[entryOff : entryOff+2])
		if tag != 0x0112 { // Orientation tag
			continue
		}

		// Type should be SHORT (3), count should be 1.
		dataType := bo.Uint16(tiff[entryOff+2 : entryOff+4])
		if dataType != 3 {
			return 0
		}

		val := int(bo.Uint16(tiff[entryOff+8 : entryOff+10]))
		if val < 1 || val > 8 {
			return 0
		}
		return val
	}

	return 0
}

// applyOrientation transforms the image according to the EXIF orientation value.
func applyOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return FlipH(img)
	case 3:
		return Rotate180(img)
	case 4:
		return FlipV(img)
	case 5:
		return Transpose(img)
	case 6:
		return Rotate270(img)
	case 7:
		return Transverse(img)
	case 8:
		return Rotate90(img)
	default:
		return img
	}
}
