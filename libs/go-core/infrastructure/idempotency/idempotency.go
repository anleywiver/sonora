// Package idempotency backs the Idempotency-Key header convention (see
// docs/api-design.md) with Redis: a repeated request with the same key
// replays the first result instead of creating a duplicate resource.
package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const ttl = 24 * time.Hour

type Store struct {
	client *redis.Client
}

func NewStore(client *redis.Client) *Store {
	return &Store{client: client}
}

func key(scope, userID, idempotencyKey string) string {
	return fmt.Sprintf("idempotency:%s:%s:%s", scope, userID, idempotencyKey)
}

// Lookup returns a previously saved value for this idempotency key, if any.
func (s *Store) Lookup(ctx context.Context, scope, userID, idempotencyKey string) (string, bool, error) {
	val, err := s.client.Get(ctx, key(scope, userID, idempotencyKey)).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

func (s *Store) Save(ctx context.Context, scope, userID, idempotencyKey, value string) error {
	return s.client.Set(ctx, key(scope, userID, idempotencyKey), value, ttl).Err()
}
