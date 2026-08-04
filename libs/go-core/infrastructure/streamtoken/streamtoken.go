// Package streamtoken issues short-lived, single-resource tokens for the
// stream endpoint, which can't carry a custom Authorization header (the
// browser's <audio> element only ever does a plain GET). Deliberately a
// separate mechanism from the main JWT access token: scoped to one song
// and a much shorter TTL (ADR 0001). Sprint 7's ws-token follows the same
// pattern.
package streamtoken

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalid = errors.New("streamtoken: invalid or expired token")

type Claims struct {
	SongID string `json:"song_id"`
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

func (i *Issuer) Issue(songID, userID uuid.UUID) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(i.ttl)
	claims := Claims{
		SongID: songID.String(),
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("streamtoken: sign: %w", err)
	}
	return signed, expiresAt, nil
}

func (i *Issuer) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalid
		}
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalid
	}
	return claims, nil
}
