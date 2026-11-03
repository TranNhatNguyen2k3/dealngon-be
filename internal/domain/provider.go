package domain

import (
	"context"
	"time"
)

type ExternalProduct struct {
	MerchantSlug   string
	Name           string
	Brand          string
	Category       string
	SKU            string
	Model          string
	Storage        string
	RAM            string
	Color          string
	Barcode        string
	Image          string
	Description    string
	Specifications map[string]string
	ProductURL     string
	Price          int64
	OriginalPrice  *int64
	Stock          *int
	Rating         *float64
	SellerName     string
	VariantName    string
}

type ExternalOffer struct {
	MerchantSlug  string
	SKU           string
	Barcode       string
	Title         string
	ProductURL    string
	Price         int64
	OriginalPrice *int64
	Stock         *int
	Rating        *float64
	SellerName    string
}

type MerchantProvider interface {
	Slug() string
	FetchProducts(ctx context.Context) ([]ExternalProduct, error)
	FetchOffers(ctx context.Context) ([]ExternalOffer, error)
}

type AffiliateProvider interface {
	GenerateLink(ctx context.Context, offer Offer) (string, error)
}

type ProductMatcher interface {
	Match(ctx context.Context, item ExternalProduct) (*Variant, error)
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	DeletePrefix(ctx context.Context, prefix string) error
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}

type TokenIssuer interface {
	Issue(adminID uint64, email string) (string, error)
	Parse(token string) (uint64, error)
}

type IPHasher interface {
	Hash(ip string) string
}
