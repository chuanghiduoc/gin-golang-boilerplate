package filetype

import (
	"bytes"
	"io"
	"mime/multipart"
)

// Magic bytes signatures for common file types
var magicSignatures = map[string][][]byte{
	"image/jpeg": {
		{0xFF, 0xD8, 0xFF},
	},
	"image/png": {
		{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	},
	"image/gif": {
		{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, // GIF87a
		{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, // GIF89a
	},
	"image/webp": {
		{0x52, 0x49, 0x46, 0x46}, // RIFF (WebP starts with RIFF)
	},
	"application/pdf": {
		{0x25, 0x50, 0x44, 0x46}, // %PDF
	},
	"application/zip": {
		{0x50, 0x4B, 0x03, 0x04}, // PK (ZIP)
		{0x50, 0x4B, 0x05, 0x06}, // Empty ZIP
		{0x50, 0x4B, 0x07, 0x08}, // Spanned ZIP
	},
	"application/msword": {
		{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, // OLE Compound Document
	},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {
		{0x50, 0x4B, 0x03, 0x04}, // DOCX is ZIP-based
	},
	"application/vnd.ms-excel": {
		{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, // OLE Compound Document
	},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {
		{0x50, 0x4B, 0x03, 0x04}, // XLSX is ZIP-based
	},
}

// Text-based types that need different validation
var textBasedTypes = map[string]bool{
	"text/plain":       true,
	"text/csv":         true,
	"application/json": true,
	"application/xml":  true,
}

// DetectMimeType reads the first bytes of a file and detects its MIME type
func DetectMimeType(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Read first 512 bytes for detection
	header := make([]byte, 512)
	n, err := f.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	return DetectFromBytes(header), nil
}

// DetectFromBytes detects MIME type from byte slice
func DetectFromBytes(data []byte) string {
	if len(data) == 0 {
		return "application/octet-stream"
	}

	// Check binary signatures first
	for mimeType, signatures := range magicSignatures {
		for _, sig := range signatures {
			if bytes.HasPrefix(data, sig) {
				// Special handling for WebP (needs additional check)
				if mimeType == "image/webp" && len(data) >= 12 {
					if !bytes.Equal(data[8:12], []byte("WEBP")) {
						continue
					}
				}
				return mimeType
			}
		}
	}

	// Check if it looks like text
	if isTextContent(data) {
		// Try to detect specific text formats
		trimmed := bytes.TrimSpace(data)
		if len(trimmed) > 0 {
			switch trimmed[0] {
			case '{', '[':
				return "application/json"
			case '<':
				if bytes.HasPrefix(trimmed, []byte("<?xml")) || bytes.HasPrefix(trimmed, []byte("<")) {
					return "application/xml"
				}
			}
		}
		return "text/plain"
	}

	return "application/octet-stream"
}

// isTextContent checks if the data appears to be text
func isTextContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Check for null bytes (binary indicator)
	for _, b := range data {
		if b == 0 {
			return false
		}
		// Check for non-printable, non-whitespace bytes
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}
	return true
}

// ValidateContent validates that file content matches the claimed MIME type
func ValidateContent(file *multipart.FileHeader, claimedType string) (bool, string, error) {
	detectedType, err := DetectMimeType(file)
	if err != nil {
		return false, "", err
	}

	// For text-based types, we're more lenient
	if textBasedTypes[claimedType] {
		if detectedType == "text/plain" || detectedType == claimedType {
			return true, detectedType, nil
		}
	}

	// For binary types, check exact match or compatible types
	if isCompatibleType(claimedType, detectedType) {
		return true, detectedType, nil
	}

	return false, detectedType, nil
}

// isCompatibleType checks if detected type is compatible with claimed type
func isCompatibleType(claimed, detected string) bool {
	if claimed == detected {
		return true
	}

	// ZIP-based formats compatibility
	zipBasedFormats := map[string]bool{
		"application/zip": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	}

	if zipBasedFormats[claimed] && detected == "application/zip" {
		return true
	}

	// OLE-based formats compatibility
	oleBasedFormats := map[string]bool{
		"application/msword":     true,
		"application/vnd.ms-excel": true,
	}

	if oleBasedFormats[claimed] && detected == "application/msword" {
		return true
	}

	return false
}
