package util

import (
	"bytes"
	"os"
	"testing"
	"unicode/utf8"
)

func ReadTest(path string, t *testing.T) string {
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("Test file does not exist: %s", err)
	}
	return SafeString(contents)
}

// returns b as a valid UTF-8 string.
func SafeString(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	var buf bytes.Buffer
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		b = b[size:]
		buf.WriteRune(r)
	}
	return buf.String()
}
