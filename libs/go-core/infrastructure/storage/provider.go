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

type QuotaInfo struct {
	// LimitBytes is 0 when the provider reports no fixed quota (rare for
	// Drive, but the field is nullable in storage_accounts either way).
	LimitBytes int64
	UsedBytes  int64
}

type Provider interface {
	// Upload writes content under filename and returns the provider's file
	// ID (e.g. a Google Drive file ID) needed to reference it later.
	Upload(ctx context.Context, filename, mimeType string, content io.Reader) (providerFileID string, err error)

	// Download fetches providerFileID's content. rangeHeader is the raw
	// HTTP Range header value from the client ("" for a full download) and
	// is forwarded as-is to the provider, which decides whether to honor it.
	Download(ctx context.Context, providerFileID, rangeHeader string) (*DownloadResult, error)

	// HealthCheck confirms the stored credentials still work and reports
	// current quota usage (Sprint 9 — quota-aware routing needs fresh
	// numbers, and a health check is the natural place to refresh them).
	HealthCheck(ctx context.Context) (*QuotaInfo, error)
}
