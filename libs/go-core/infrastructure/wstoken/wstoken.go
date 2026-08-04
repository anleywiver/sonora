// Package wstoken issues short-lived, single-use tokens for the WebSocket
// handshake (`ws://.../ws?token=`) — the handshake can't send a custom
// Authorization header, and unlike the stream token (Sprint 4, one token
// per song, replayable within its TTL), a ws-token is consumed exactly
// once: reusing a leaked token to open a second connection must fail.
package wstoken

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrInvalid = errors.New("wstoken: invalid, expired, or already used token")

const ttl = 60 * time.Second

type Issuer struct {
	client *redis.Client
}

func NewIssuer(client *redis.Client) *Issuer {
	return &Issuer{client: client}
}

type claims struct {
	UserID string `json:"user_id"`
}

func (i *Issuer) Issue(ctx context.Context, userID uuid.UUID) (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("wstoken: generate: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buf)

	payload, err := json.Marshal(claims{UserID: userID.String()})
	if err != nil {
		return "", fmt.Errorf("wstoken: marshal: %w", err)
	}
	if err := i.client.Set(ctx, key(token), payload, ttl).Err(); err != nil {
		return "", fmt.Errorf("wstoken: store: %w", err)
	}
	return token, nil
}

// Consume validates the token and atomically deletes it (GETDEL) so two
// connection attempts racing on the same token can't both succeed.
func (i *Issuer) Consume(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := i.client.GetDel(ctx, key(token)).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.UUID{}, ErrInvalid
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("wstoken: consume: %w", err)
	}

	var c claims
	if err := json.Unmarshal([]byte(val), &c); err != nil {
		return uuid.UUID{}, fmt.Errorf("wstoken: unmarshal: %w", err)
	}
	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return uuid.UUID{}, ErrInvalid
	}
	return userID, nil
}

func key(token string) string {
	return "wstoken:" + token
}
