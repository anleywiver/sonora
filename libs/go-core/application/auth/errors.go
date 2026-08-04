package auth

import "errors"

var ErrInvalidRefreshToken = errors.New("auth: invalid or expired refresh token")

// ErrRefreshTokenReused signals that an already-rotated-out refresh token
// was presented again — every session for that user has just been
// revoked as a precaution (Sprint 12, ADR 0006). The caller should treat
// this the same as ErrInvalidRefreshToken for the response (401,
// generic message) — the distinction exists for logging, not for the API
// contract.
var ErrRefreshTokenReused = errors.New("auth: refresh token reuse detected, all sessions revoked")
