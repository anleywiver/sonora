package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GoogleDriveProvider uploads to a single Drive account via a stored OAuth
// refresh token, scoped to drive.file (only files this app creates —
// least privilege, never broad Drive read/write access).
type GoogleDriveProvider struct {
	tokenSource oauth2.TokenSource
}

func NewGoogleDriveProvider(ctx context.Context, clientID, clientSecret, refreshToken string) *GoogleDriveProvider {
	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     googleoauth.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/drive.file"},
	}
	token := &oauth2.Token{RefreshToken: refreshToken}
	return &GoogleDriveProvider{tokenSource: conf.TokenSource(ctx, token)}
}

func (p *GoogleDriveProvider) Upload(ctx context.Context, filename, mimeType string, content io.Reader) (string, error) {
	svc, err := drive.NewService(ctx, option.WithTokenSource(p.tokenSource))
	if err != nil {
		return "", fmt.Errorf("storage: drive client: %w", err)
	}
	file := &drive.File{Name: filename}
	created, err := svc.Files.Create(file).Media(content).Fields("id").Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("storage: drive upload: %w", err)
	}
	return created.Id, nil
}

// Download forwards rangeHeader as-is to Drive's media download endpoint,
// which honors it the same way any HTTP file server would (206 + a real
// Content-Range on success). Needed so the stream handler can support
// seeking in the browser's <audio> element.
func (p *GoogleDriveProvider) Download(ctx context.Context, providerFileID, rangeHeader string) (*DownloadResult, error) {
	svc, err := drive.NewService(ctx, option.WithTokenSource(p.tokenSource))
	if err != nil {
		return nil, fmt.Errorf("storage: drive client: %w", err)
	}

	call := svc.Files.Get(providerFileID).Context(ctx)
	if rangeHeader != "" {
		call.Header().Set("Range", rangeHeader)
	}
	resp, err := call.Download()
	if err != nil {
		return nil, fmt.Errorf("storage: drive download: %w", err)
	}

	return &DownloadResult{
		Body:          resp.Body,
		ContentLength: resp.ContentLength,
		ContentRange:  resp.Header.Get("Content-Range"),
		Partial:       resp.StatusCode == http.StatusPartialContent,
	}, nil
}

// HealthCheck calls Drive's About.get, which both confirms the refresh
// token still works and returns current quota usage in one round trip.
func (p *GoogleDriveProvider) HealthCheck(ctx context.Context) (*QuotaInfo, error) {
	svc, err := drive.NewService(ctx, option.WithTokenSource(p.tokenSource))
	if err != nil {
		return nil, fmt.Errorf("storage: drive client: %w", err)
	}
	about, err := svc.About.Get().Fields("storageQuota").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("storage: drive about: %w", err)
	}
	if about.StorageQuota == nil {
		return nil, fmt.Errorf("storage: drive about: no storageQuota in response")
	}
	return &QuotaInfo{
		LimitBytes: about.StorageQuota.Limit,
		UsedBytes:  about.StorageQuota.Usage,
	}, nil
}
