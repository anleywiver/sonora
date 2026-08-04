// Package redis provides the shared go-redis client used for caching,
// playback state, and idempotency keys (Asynq manages its own connection
// separately via asynq.RedisClientOpt).
package redis

import "github.com/redis/go-redis/v9"

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}
