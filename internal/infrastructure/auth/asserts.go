package auth

import "github.com/dealngon/backend/internal/domain"

var (
	_ domain.TokenIssuer = (*Issuer)(nil)
	_ domain.IPHasher    = (*Issuer)(nil)
)
