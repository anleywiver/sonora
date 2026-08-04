// Package meilisearch wraps the official SDK for the "songs" search index.
// Meilisearch auto-creates the index on the first document add, but only
// if it can infer a single primary-key candidate — our documents have
// several fields ending in "id" (id, artist_id, album_id), so primaryKey
// must be passed explicitly or index creation fails outright.
package meilisearch

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	meili "github.com/meilisearch/meilisearch-go"
)

const SongsIndex = "songs"

var songsPrimaryKey = meili.StringPtr("id")

type Client struct {
	svc meili.ServiceManager
}

func NewClient(host, apiKey string) *Client {
	return &Client{svc: meili.New(host, meili.WithAPIKey(apiKey))}
}

type SongDocument struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ArtistID   string `json:"artist_id"`
	ArtistName string `json:"artist_name"`
	AlbumID    string `json:"album_id,omitempty"`
	AlbumTitle string `json:"album_title,omitempty"`
	DurationMs int    `json:"duration_ms"`
}

func (c *Client) IndexSong(ctx context.Context, doc SongDocument) error {
	opts := &meili.DocumentOptions{PrimaryKey: songsPrimaryKey}
	if _, err := c.svc.Index(SongsIndex).AddDocumentsWithContext(ctx, []SongDocument{doc}, opts); err != nil {
		return fmt.Errorf("meilisearch: index song: %w", err)
	}
	return nil
}

type SearchResult struct {
	Hits           []SongDocument
	EstimatedTotal int64
}

func (c *Client) SearchSongs(ctx context.Context, query string, limit int64) (*SearchResult, error) {
	resp, err := c.svc.Index(SongsIndex).SearchWithContext(ctx, query, &meili.SearchRequest{Limit: limit})
	if err != nil {
		// A fresh install (or right after any full reindex) has no songs
		// index at all yet — that's "no results", not a search failure.
		var meiliErr *meili.Error
		if errors.As(err, &meiliErr) && meiliErr.StatusCode == http.StatusNotFound {
			return &SearchResult{}, nil
		}
		return nil, fmt.Errorf("meilisearch: search: %w", err)
	}
	var docs []SongDocument
	if err := resp.Hits.Decode(&docs); err != nil {
		return nil, fmt.Errorf("meilisearch: decode hits: %w", err)
	}
	return &SearchResult{Hits: docs, EstimatedTotal: resp.EstimatedTotalHits}, nil
}
