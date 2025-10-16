package helper

import (
	"bytes"
	"net/url"
)

// EncodeURIComponent encodes a string equivalent to JavaScript's encodeURIComponent.
// It keeps only: A–Z a–z 0–9 - _ . ! ~ * ' ( )
// All other bytes (UTF-8) are percent-encoded as %HH (uppercase).
func EncodeURIComponent(s string) string {
	var buf bytes.Buffer
	const hex = "0123456789ABCDEF"

	for _, b := range []byte(s) { // process UTF-8 bytes directly
		// unescaped set: A–Z a–z 0–9 - _ . ! ~ * ' ( )
		if (b >= 'A' && b <= 'Z') ||
			(b >= 'a' && b <= 'z') ||
			(b >= '0' && b <= '9') ||
			b == '-' || b == '_' || b == '.' ||
			b == '!' || b == '~' || b == '*' ||
			b == '\'' || b == '(' || b == ')' {
			buf.WriteByte(b)
			continue
		}
		buf.WriteByte('%')
		buf.WriteByte(hex[b>>4])
		buf.WriteByte(hex[b&0x0F])
	}
	return buf.String()
}

// DecodeURIComponent decodes a string produced by EncodeURIComponent.
func DecodeURIComponent(s string) (string, error) {
	// encodeURIComponent never produces '+', so a simple percent-unescape is enough.
	// Reuse the standard library for correctness.
	// Note: importing net/url only here keeps the public API small.
	return url.PathUnescape(s)
}
