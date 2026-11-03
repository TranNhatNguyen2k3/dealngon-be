package rediscache

import "github.com/dealngon/backend/internal/domain"

var (
	_ domain.Cache       = (*Client)(nil)
	_ domain.RateLimiter = (*Client)(nil)
)
