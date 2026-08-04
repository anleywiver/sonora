package ingest

import "errors"

var (
	ErrNotFound               = errors.New("ingest: job not found")
	ErrNoActiveStorageAccount = errors.New("ingest: no active storage account configured")
	ErrNotRetryable           = errors.New("ingest: job is not in a retryable state")
)
