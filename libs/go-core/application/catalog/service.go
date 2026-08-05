// Package catalog implements read access to songs/albums/artists/genres,
// and the stream pipeline: issue a short-lived scoped token, then fetch
// the song's bytes from the storage provider with Range support.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"sonora.dev/go-core/infrastructure/crypto"
	"sonora.dev/go-core/infrastructure/postgres/sqlc"
	"sonora.dev/go-core/infrastructure/storage"
	"sonora.dev/go-core/infrastructure/streamtoken"
)

var ErrNotFound = errors.New("catalog: not found")

const streamTokenTTL = 5 * time.Minute

type Service struct {
	q                  *sqlc.Queries
	box                *crypto.Box
	streamTokens       *streamtoken.Issuer
	googleClientID     string
	googleClientSecret string
}

func NewService(q *sqlc.Queries, box *crypto.Box, streamTokenSecret, googleClientID, googleClientSecret string) *Service {
	return &Service{
		q:                  q,
		box:                box,
		streamTokens:       streamtoken.NewIssuer(streamTokenSecret, streamTokenTTL),
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
	}
}

type SongDetail struct {
	ID               uuid.UUID
	Title            string
	DurationMs       int
	TrackNumber      *int
	ArtistID         uuid.UUID
	ArtistName       string
	AlbumID          *uuid.UUID
	AlbumTitle       string
	AlbumCoverURL    string
	StorageFileID    uuid.UUID
	ProviderFileID   string
	MimeType         string
	SizeBytes        int64
	StorageAccountID uuid.UUID
}

func (s *Service) GetSong(ctx context.Context, id uuid.UUID) (*SongDetail, error) {
	row, err := s.q.GetSongDetail(ctx, toPgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("catalog: get song: %w", err)
	}

	var trackNumber *int
	if row.TrackNumber != nil {
		n := int(*row.TrackNumber)
		trackNumber = &n
	}

	return &SongDetail{
		ID:               fromPgUUID(row.ID),
		Title:            row.Title,
		DurationMs:       int(row.DurationMs),
		TrackNumber:      trackNumber,
		ArtistID:         fromPgUUID(row.ArtistID),
		ArtistName:       row.ArtistName,
		AlbumID:          fromPgUUIDPtr(row.AlbumID),
		AlbumTitle:       strOrEmpty(row.AlbumTitle),
		AlbumCoverURL:    strOrEmpty(row.AlbumCoverUrl),
		StorageFileID:    fromPgUUID(row.StorageFileID),
		ProviderFileID:   row.ProviderFileID,
		MimeType:         row.MimeType,
		SizeBytes:        row.SizeBytes,
		StorageAccountID: fromPgUUID(row.StorageAccountID),
	}, nil
}

type AlbumDetail struct {
	ID         uuid.UUID
	Title      string
	CoverURL   string
	ArtistID   uuid.UUID
	ArtistName string
	ReleasedAt *time.Time
}

func (s *Service) GetAlbum(ctx context.Context, id uuid.UUID) (*AlbumDetail, error) {
	row, err := s.q.GetAlbumDetail(ctx, toPgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("catalog: get album: %w", err)
	}
	var releasedAt *time.Time
	if row.ReleasedAt.Valid {
		releasedAt = &row.ReleasedAt.Time
	}
	return &AlbumDetail{
		ID:         fromPgUUID(row.ID),
		Title:      row.Title,
		CoverURL:   strOrEmpty(row.CoverUrl),
		ArtistID:   fromPgUUID(row.ArtistID),
		ArtistName: row.ArtistName,
		ReleasedAt: releasedAt,
	}, nil
}

type Song struct {
	ID          uuid.UUID
	Title       string
	DurationMs  int
	TrackNumber *int
	AlbumID     *uuid.UUID
	// ArtistName is only populated by ListRecent (Home's Trending widget
	// needs it since, unlike ListSongsByAlbum/ListSongsByArtist, results
	// aren't already scoped to one artist) — "" for other callers.
	ArtistName string
}

func (s *Service) ListSongsByAlbum(ctx context.Context, albumID uuid.UUID) ([]*Song, error) {
	rows, err := s.q.ListSongsByAlbum(ctx, toPgUUID(albumID))
	if err != nil {
		return nil, fmt.Errorf("catalog: list songs by album: %w", err)
	}
	out := make([]*Song, 0, len(rows))
	for _, row := range rows {
		var trackNumber *int
		if row.TrackNumber != nil {
			n := int(*row.TrackNumber)
			trackNumber = &n
		}
		out = append(out, &Song{
			ID:          fromPgUUID(row.ID),
			Title:       row.Title,
			DurationMs:  int(row.DurationMs),
			TrackNumber: trackNumber,
			AlbumID:     fromPgUUIDPtr(row.AlbumID),
		})
	}
	return out, nil
}

const popularSongsLimit = 10

// ListSongsByArtist stands in for "Popular songs" on Artist Detail (most
// recently added — same "no real play-count ranking yet" caveat as
// ListRecent above; play_history is per-user, not a global per-song
// counter we could rank by without a new aggregate query).
func (s *Service) ListSongsByArtist(ctx context.Context, artistID uuid.UUID) ([]*Song, error) {
	rows, err := s.q.ListSongsByArtist(ctx, sqlc.ListSongsByArtistParams{
		ArtistID: toPgUUID(artistID),
		Limit:    popularSongsLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("catalog: list songs by artist: %w", err)
	}
	out := make([]*Song, 0, len(rows))
	for _, row := range rows {
		var trackNumber *int
		if row.TrackNumber != nil {
			n := int(*row.TrackNumber)
			trackNumber = &n
		}
		out = append(out, &Song{
			ID:          fromPgUUID(row.ID),
			Title:       row.Title,
			DurationMs:  int(row.DurationMs),
			TrackNumber: trackNumber,
			AlbumID:     fromPgUUIDPtr(row.AlbumID),
		})
	}
	return out, nil
}

type Artist struct {
	ID       uuid.UUID
	Name     string
	ImageURL string
}

func (s *Service) GetArtist(ctx context.Context, id uuid.UUID) (*Artist, error) {
	row, err := s.q.GetArtistByID(ctx, toPgUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("catalog: get artist: %w", err)
	}
	return &Artist{ID: fromPgUUID(row.ID), Name: row.Name, ImageURL: strOrEmpty(row.ImageUrl)}, nil
}

type Album struct {
	ID       uuid.UUID
	Title    string
	CoverURL string
}

func (s *Service) ListAlbumsByArtist(ctx context.Context, artistID uuid.UUID) ([]*Album, error) {
	rows, err := s.q.ListAlbumsByArtist(ctx, toPgUUID(artistID))
	if err != nil {
		return nil, fmt.Errorf("catalog: list albums by artist: %w", err)
	}
	out := make([]*Album, 0, len(rows))
	for _, row := range rows {
		out = append(out, &Album{ID: fromPgUUID(row.ID), Title: row.Title, CoverURL: strOrEmpty(row.CoverUrl)})
	}
	return out, nil
}

type Genre struct {
	ID   uuid.UUID
	Name string
}

func (s *Service) ListGenres(ctx context.Context) ([]*Genre, error) {
	rows, err := s.q.ListGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("catalog: list genres: %w", err)
	}
	out := make([]*Genre, 0, len(rows))
	for _, row := range rows {
		out = append(out, &Genre{ID: fromPgUUID(row.ID), Name: row.Name})
	}
	return out, nil
}

// ListRecent stands in for "trending" until play_history exists (Sprint 6)
// gives us real play-count data to rank by.
func (s *Service) ListRecent(ctx context.Context, limit int32) ([]*Song, error) {
	rows, err := s.q.ListRecentSongs(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("catalog: list recent songs: %w", err)
	}
	out := make([]*Song, 0, len(rows))
	for _, row := range rows {
		out = append(out, &Song{
			ID:         fromPgUUID(row.ID),
			Title:      row.Title,
			DurationMs: int(row.DurationMs),
			AlbumID:    fromPgUUIDPtr(row.AlbumID),
			ArtistName: row.ArtistName,
		})
	}
	return out, nil
}

const libraryBrowseLimit = 200

// LibrarySong/LibraryAlbum/LibraryArtist back the Sprint 14 sisipan
// Browse Library page (ADR 0011) — the whole catalog, not just favorites.

type LibrarySong struct {
	ID         uuid.UUID
	Title      string
	DurationMs int
	ArtistName string
	AlbumTitle string
}

func (s *Service) ListLibrarySongs(ctx context.Context, search string, sortAlpha bool) ([]*LibrarySong, error) {
	params := sqlc.ListLibrarySongsParams{SortAlpha: sortAlpha, LimitCount: libraryBrowseLimit}
	if search != "" {
		params.Search = &search
	}
	rows, err := s.q.ListLibrarySongs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("catalog: list library songs: %w", err)
	}
	out := make([]*LibrarySong, 0, len(rows))
	for _, row := range rows {
		out = append(out, &LibrarySong{
			ID: fromPgUUID(row.ID), Title: row.Title, DurationMs: int(row.DurationMs),
			ArtistName: row.ArtistName, AlbumTitle: strOrEmpty(row.AlbumTitle),
		})
	}
	return out, nil
}

type LibraryAlbum struct {
	ID         uuid.UUID
	Title      string
	CoverURL   string
	ArtistName string
}

func (s *Service) ListLibraryAlbums(ctx context.Context, search string, sortAlpha bool) ([]*LibraryAlbum, error) {
	params := sqlc.ListLibraryAlbumsParams{SortAlpha: sortAlpha, LimitCount: libraryBrowseLimit}
	if search != "" {
		params.Search = &search
	}
	rows, err := s.q.ListLibraryAlbums(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("catalog: list library albums: %w", err)
	}
	out := make([]*LibraryAlbum, 0, len(rows))
	for _, row := range rows {
		out = append(out, &LibraryAlbum{
			ID: fromPgUUID(row.ID), Title: row.Title, CoverURL: strOrEmpty(row.CoverUrl), ArtistName: row.ArtistName,
		})
	}
	return out, nil
}

type LibraryArtist struct {
	ID       uuid.UUID
	Name     string
	ImageURL string
}

func (s *Service) ListLibraryArtists(ctx context.Context, search string, sortAlpha bool) ([]*LibraryArtist, error) {
	params := sqlc.ListLibraryArtistsParams{SortAlpha: sortAlpha, LimitCount: libraryBrowseLimit}
	if search != "" {
		params.Search = &search
	}
	rows, err := s.q.ListLibraryArtists(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("catalog: list library artists: %w", err)
	}
	out := make([]*LibraryArtist, 0, len(rows))
	for _, row := range rows {
		out = append(out, &LibraryArtist{ID: fromPgUUID(row.ID), Name: row.Name, ImageURL: strOrEmpty(row.ImageUrl)})
	}
	return out, nil
}

// IssueStreamToken verifies the song exists, then issues a 5-minute token
// scoped to it — used as a query param on GET /songs/:id/stream since
// <audio> can't send a custom Authorization header (ADR 0001).
func (s *Service) IssueStreamToken(ctx context.Context, songID, userID uuid.UUID) (token string, expiresAt time.Time, err error) {
	if _, err := s.GetSong(ctx, songID); err != nil {
		return "", time.Time{}, err
	}
	return s.streamTokens.Issue(songID, userID)
}

// ParseStreamToken validates the token and confirms it was issued for
// songID specifically — a token for one song can't be replayed on another.
func (s *Service) ParseStreamToken(tokenString string, songID uuid.UUID) (userID uuid.UUID, err error) {
	claims, err := s.streamTokens.Parse(tokenString)
	if err != nil {
		return uuid.UUID{}, streamtoken.ErrInvalid
	}
	if claims.SongID != songID.String() {
		return uuid.UUID{}, streamtoken.ErrInvalid
	}
	parsedUserID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.UUID{}, streamtoken.ErrInvalid
	}
	return parsedUserID, nil
}

type StreamResult struct {
	*storage.DownloadResult
	MimeType string
}

// Stream fetches the song's bytes from its storage provider, forwarding
// rangeHeader so the browser can seek.
func (s *Service) Stream(ctx context.Context, songID uuid.UUID, rangeHeader string) (*StreamResult, error) {
	song, err := s.GetSong(ctx, songID)
	if err != nil {
		return nil, err
	}

	account, err := s.q.GetStorageAccountByID(ctx, toPgUUID(song.StorageAccountID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("catalog: load storage account: %w", err)
	}

	refreshToken, err := s.box.Decrypt(account.CredentialsEncrypted)
	if err != nil {
		return nil, fmt.Errorf("catalog: decrypt storage credentials: %w", err)
	}

	provider := storage.NewGoogleDriveProvider(ctx, s.googleClientID, s.googleClientSecret, refreshToken)
	result, err := provider.Download(ctx, song.ProviderFileID, rangeHeader)
	if err != nil {
		return nil, fmt.Errorf("catalog: download from storage: %w", err)
	}
	return &StreamResult{DownloadResult: result, MimeType: song.MimeType}, nil
}
