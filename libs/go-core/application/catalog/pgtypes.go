package catalog

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func fromPgUUID(id pgtype.UUID) uuid.UUID {
	return uuid.UUID(id.Bytes)
}

func fromPgUUIDPtr(id pgtype.UUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}
	v := fromPgUUID(id)
	return &v
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
