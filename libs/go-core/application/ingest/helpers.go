package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func detectMimeType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

// originalFilenameFrom recovers the user's original filename (sans
// extension) from the temp file basename, which the HTTP handler writes as
// "<random>__<sanitized-original-name>". Used as a title fallback when the
// audio file has no title tag.
func originalFilenameFrom(tempPath string) string {
	base := filepath.Base(tempPath)
	parts := strings.SplitN(base, "__", 2)
	name := base
	if len(parts) == 2 {
		name = parts[1]
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}
