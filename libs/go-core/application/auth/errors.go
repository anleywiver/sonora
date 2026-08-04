package auth

import "errors"

var ErrInvalidRefreshToken = errors.New("auth: invalid or expired refresh token")
