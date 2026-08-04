// Package storage abstracts the backing object store. Google Drive is the
// only implementation for now; Hetzner Storage Box slots in behind the same
// interface later (ADR 0001) without touching callers.
package storage

import (
	"context"
	"io"
)

type Provider interface {
	// Upload writes content under filename and returns the provider's file
	// ID (e.g. a Google Drive file ID) needed to reference it later.
	Upload(ctx context.Context, filename, mimeType string, content io.Reader) (providerFileID string, err error)
}
