package storage

import (
	"context"
	"fmt"
	"io"

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
