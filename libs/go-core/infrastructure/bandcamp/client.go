// Package bandcamp lists and downloads a fan's own paid purchases via
// Bandcamp's fancollection API — the same endpoint the official Bandcamp
// app uses for "redownload my collection", not a scrape of content the
// user doesn't own. See docs/decisions/0004-sprint10-scheduled-jobs-and-ingest-sources.md.
//
// There is no public OAuth for this; the credential is a session cookie
// the user copies out of their own logged-in browser (same manual/
// out-of-band pattern as the Drive refresh token in ADR 0002).
package bandcamp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const collectionItemsURL = "https://bandcamp.com/api/fancollection/1/collection_items"

// ErrAlbumNotSupported is returned by Download for a multi-track album
// purchase (delivered as a zip) — see the "batasan scope v1" note in ADR
// 0004. Callers should log and skip, not fail the whole sync.
var ErrAlbumNotSupported = errors.New("bandcamp: album (zip) downloads are not supported yet, only single-track purchases")

type Credentials struct {
	IdentityCookie string `json:"identity_cookie"`
	FanID          string `json:"fan_id"`
}

type Purchase struct {
	Token       string
	ItemType    string // "track" or "album"
	Title       string
	ArtistName  string
	PurchasedAt time.Time
	// RedownloadURL is the collection page Download resolves to find the
	// real, signed file URL — it is NOT the file itself.
	RedownloadURL string
}

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

type collectionItemsRequest struct {
	FanID          string `json:"fan_id"`
	OlderThanToken string `json:"older_than_token"`
	Count          int    `json:"count"`
}

type collectionItemsResponse struct {
	Items []struct {
		ItemType   string `json:"item_type"`
		ItemTitle  string `json:"item_title"`
		BandName   string `json:"band_name"`
		Purchased  string `json:"purchased"`
		Token      string `json:"token"`
		SaleItemID int64  `json:"sale_item_id"`
	} `json:"items"`
	RedownloadURLs map[string]string `json:"redownload_urls"`
}

// ListNewPurchases returns purchases made after since, newest first
// (Bandcamp's own pagination order). Only the first page is fetched
// (count: 100) — a personal collection realistically never needs more
// than one page per sync interval.
func (c *Client) ListNewPurchases(ctx context.Context, creds Credentials, since time.Time) ([]Purchase, error) {
	reqBody, err := json.Marshal(collectionItemsRequest{
		FanID:          creds.FanID,
		OlderThanToken: fmt.Sprintf("%d:0:a::", time.Now().Unix()),
		Count:          100,
	})
	if err != nil {
		return nil, fmt.Errorf("bandcamp: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, collectionItemsURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("bandcamp: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "identity="+creds.IdentityCookie)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bandcamp: collection_items request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bandcamp: collection_items returned %d (check identity cookie is still valid)", resp.StatusCode)
	}

	var parsed collectionItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("bandcamp: decode collection_items response: %w", err)
	}

	purchases := make([]Purchase, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		purchasedAt, err := time.Parse(time.RFC3339, item.Purchased)
		if err != nil {
			continue
		}
		if !purchasedAt.After(since) {
			continue
		}
		redownloadURL := parsed.RedownloadURLs[fmt.Sprintf("p%d", item.SaleItemID)]
		purchases = append(purchases, Purchase{
			Token:         item.Token,
			ItemType:      item.ItemType,
			Title:         item.ItemTitle,
			ArtistName:    item.BandName,
			PurchasedAt:   purchasedAt,
			RedownloadURL: redownloadURL,
		})
	}
	return purchases, nil
}

var dataBlobPattern = regexp.MustCompile(`data-blob="([^"]+)"`)

type downloadBlob struct {
	DownloadItems []struct {
		Downloads map[string]struct {
			URL string `json:"url"`
		} `json:"downloads"`
	} `json:"download_items"`
}

// Download resolves purchase's redownload page to a real file URL and
// streams it to w. Only single-track purchases are supported (see
// ErrAlbumNotSupported).
func (c *Client) Download(ctx context.Context, purchase Purchase, w io.Writer) error {
	if purchase.ItemType != "track" {
		return ErrAlbumNotSupported
	}
	if purchase.RedownloadURL == "" {
		return errors.New("bandcamp: purchase has no redownload URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, purchase.RedownloadURL, nil)
	if err != nil {
		return fmt.Errorf("bandcamp: build redownload page request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("bandcamp: fetch redownload page: %w", err)
	}
	defer resp.Body.Close()

	page, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("bandcamp: read redownload page: %w", err)
	}

	matches := dataBlobPattern.FindSubmatch(page)
	if matches == nil {
		return errors.New("bandcamp: redownload page missing data-blob (page format may have changed)")
	}

	var blob downloadBlob
	if err := json.Unmarshal(unescapeHTMLAttr(matches[1]), &blob); err != nil {
		return fmt.Errorf("bandcamp: parse download blob: %w", err)
	}
	if len(blob.DownloadItems) == 0 {
		return errors.New("bandcamp: no download items in blob")
	}

	downloads := blob.DownloadItems[0].Downloads
	format, ok := downloads["mp3-320"]
	if !ok {
		for _, d := range downloads {
			format = d
			ok = true
			break
		}
	}
	if !ok || format.URL == "" {
		return errors.New("bandcamp: no downloadable format found")
	}

	fileReq, err := http.NewRequestWithContext(ctx, http.MethodGet, format.URL, nil)
	if err != nil {
		return fmt.Errorf("bandcamp: build file request: %w", err)
	}
	fileResp, err := c.httpClient.Do(fileReq)
	if err != nil {
		return fmt.Errorf("bandcamp: download file: %w", err)
	}
	defer fileResp.Body.Close()

	if _, err := io.Copy(w, fileResp.Body); err != nil {
		return fmt.Errorf("bandcamp: write file: %w", err)
	}
	return nil
}

// unescapeHTMLAttr undoes the HTML entity escaping Bandcamp applies to
// the data-blob attribute value (only the handful of entities that can
// legally appear inside a double-quoted attribute).
func unescapeHTMLAttr(b []byte) []byte {
	s := string(b)
	s = strings.ReplaceAll(s, "&quot;", `"`)
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return []byte(s)
}
