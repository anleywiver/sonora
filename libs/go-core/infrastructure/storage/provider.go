// Package storage abstracts the backing object store. Google Drive is the
// only implementation for now; Hetzner Storage Box slots in behind the same
// interface later (ADR 0001) without touching callers.
package storage

import (
	"context"
	"io"
)

type DownloadResult struct {
	Body          io.ReadCloser
	ContentLength int64
	// ContentRange is the provider's Content-Range response header value,
	// set only when Partial is true.
	ContentRange string
	Partial      bool
}

type Provider interface {
	// Upload writes content under filename and returns the provider's file
	// ID (e.g. a Google Drive file ID) needed to reference it later.
	Upload(ctx context.Context, filename, mimeType string, content io.Reader) (providerFileID string, err error)

	// Download fetches providerFileID's content. rangeHeader is the raw
	// HTTP Range header value from the client ("" for a full download) and
	// is forwarded as-is to the provider, which decides whether to honor it.
	Download(ctx context.Context, providerFileID, rangeHeader string) (*DownloadResult, error)
}
