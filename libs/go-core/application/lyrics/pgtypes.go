package lyrics

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
