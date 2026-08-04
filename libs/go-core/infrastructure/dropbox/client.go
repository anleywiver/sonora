// Package dropbox implements the "cloud_sync" ingest source (ADR 0004,
// decision 5): watch a folder in the user's own Dropbox for new audio
// files and ingest them. OneDrive/iCloud are explicitly out of scope for
// now — same source_type, different client can slot in later without a
// schema change.
package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	listFolderURL = "https://api.dropboxapi.com/2/files/list_folder"
	downloadURL   = "https://content.dropboxapi.com/2/files/download"
	tokenURL      = "https://api.dropboxapi.com/oauth2/token"
)

var audioExtensions = []string{".mp3", ".flac", ".m4a", ".wav", ".ogg"}

type Credentials struct {
	RefreshToken string `json:"refresh_token"`
	AppKey       string `json:"app_key"`
	AppSecret    string `json:"app_secret"`
	FolderPath   string `json:"folder_path"`
}

type File struct {
	Path           string
	Name           string
	ServerModified time.Time
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 60 * time.Second}}
}

func (c *Client) tokenSource(ctx context.Context, creds Credentials) oauth2.TokenSource {
	conf := &oauth2.Config{
		ClientID:     creds.AppKey,
		ClientSecret: creds.AppSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: tokenURL},
	}
	token := &oauth2.Token{RefreshToken: creds.RefreshToken}
	return conf.TokenSource(ctx, token)
}

type listFolderRequest struct {
	Path string `json:"path"`
}

type listFolderResponse struct {
	Entries []struct {
		Tag            string `json:".tag"`
		Name           string `json:"name"`
		PathLower      string `json:"path_lower"`
		ServerModified string `json:"server_modified"`
	} `json:"entries"`
}

// ListNewFiles returns audio files in the configured folder modified
// after since. Only the first page is fetched (has_more/cursor pagination
// isn't implemented — a personal sync folder realistically fits in one
// page; revisit if that stops being true).
func (c *Client) ListNewFiles(ctx context.Context, creds Credentials, since time.Time) ([]File, error) {
	token, err := c.tokenSource(ctx, creds).Token()
	if err != nil {
		return nil, fmt.Errorf("dropbox: refresh access token: %w", err)
	}

	reqBody, err := json.Marshal(listFolderRequest{Path: creds.FolderPath})
	if err != nil {
		return nil, fmt.Errorf("dropbox: marshal list_folder request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, listFolderURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, fmt.Errorf("dropbox: build list_folder request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dropbox: list_folder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("dropbox: list_folder returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed listFolderResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("dropbox: decode list_folder response: %w", err)
	}

	files := make([]File, 0, len(parsed.Entries))
	for _, entry := range parsed.Entries {
		if entry.Tag != "file" || !hasAudioExtension(entry.Name) {
			continue
		}
		modified, err := time.Parse(time.RFC3339, entry.ServerModified)
		if err != nil || !modified.After(since) {
			continue
		}
		files = append(files, File{Path: entry.PathLower, Name: entry.Name, ServerModified: modified})
	}
	return files, nil
}

type downloadAPIArg struct {
	Path string `json:"path"`
}

// Download streams file's content to w.
func (c *Client) Download(ctx context.Context, creds Credentials, file File, w io.Writer) error {
	token, err := c.tokenSource(ctx, creds).Token()
	if err != nil {
		return fmt.Errorf("dropbox: refresh access token: %w", err)
	}

	argJSON, err := json.Marshal(downloadAPIArg{Path: file.Path})
	if err != nil {
		return fmt.Errorf("dropbox: marshal download arg: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("dropbox: build download request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Dropbox-API-Arg", string(argJSON))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox: download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("dropbox: download returned %d: %s", resp.StatusCode, string(body))
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("dropbox: write file: %w", err)
	}
	return nil
}

func hasAudioExtension(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range audioExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
