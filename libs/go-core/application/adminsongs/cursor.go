package adminsongs

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Same opaque cursor shape used across this API (e.g. application/ingest)
// — base64 of "<created_at_unixnano>_<id>", keyset pagination.
func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%d_%s", createdAt.UnixNano(), id.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, uuid.UUID{}, fmt.Errorf("adminsongs: invalid cursor: %w", err)
	}
	parts := strings.SplitN(string(raw), "_", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.UUID{}, fmt.Errorf("adminsongs: invalid cursor format")
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, uuid.UUID{}, fmt.Errorf("adminsongs: invalid cursor timestamp: %w", err)
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.UUID{}, fmt.Errorf("adminsongs: invalid cursor id: %w", err)
	}
	return time.Unix(0, nanos), id, nil
}
