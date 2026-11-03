package domain

import (
	"context"
	"time"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id uint64) (*Product, error)
	FindBySlug(ctx context.Context, slug string) (*Product, error)
	Search(ctx context.Context, query SearchQuery) ([]ProductCard, int64, error)
	List(ctx context.Context, page, limit int) ([]Product, int64, error)
	Create(ctx context.Context, product Product) (*Product, error)
	Update(ctx context.Context, product Product) (*Product, error)
	Delete(ctx context.Context, id uint64) error
	Count(ctx context.Context) (int64, error)
	ListIDs(ctx context.Context) ([]uint64, error)
	ListSitemap(ctx context.Context) ([]SitemapEntry, error)
}

type VariantRepository interface {
	ListByProduct(ctx context.Context, productID uint64) ([]Variant, error)
	FindBySKU(ctx context.Context, sku string) (*Variant, error)
	FindByBarcode(ctx context.Context, barcode string) (*Variant, error)
	FindByBrandModel(ctx context.Context, brand, model string) (*Variant, error)
	FindByAttributes(ctx context.Context, brand, model, storage, ram, color string) (*Variant, error)
	FindByNormalizedName(ctx context.Context, name string) (*Variant, error)
	Upsert(ctx context.Context, variant Variant) (*Variant, error)
	Update(ctx context.Context, variant Variant) (*Variant, error)
}

type OfferRepository interface {
	FindByID(ctx context.Context, id uint64) (*Offer, error)
	ListByProduct(ctx context.Context, productID uint64) ([]Offer, error)
	List(ctx context.Context, page, limit int) ([]Offer, int64, error)
	Upsert(ctx context.Context, offer Offer) (*Offer, error)
	Update(ctx context.Context, offer Offer) (*Offer, error)
	Delete(ctx context.Context, id uint64) error
	Count(ctx context.Context) (int64, error)
	SetAffiliateURL(ctx context.Context, offerID uint64, affiliateURL string) error
}

type PriceHistoryRepository interface {
	ListByProduct(ctx context.Context, productID uint64) ([]PricePoint, error)
	ListByOffer(ctx context.Context, offerID uint64) ([]PricePoint, error)
	Backfill(ctx context.Context, offerID uint64, points []PricePoint) error
	AppendIfChanged(ctx context.Context, offerID uint64, price int64, at time.Time) error
	HasHistory(ctx context.Context, offerID uint64) (bool, error)
}

type DealRepository interface {
	List(ctx context.Context, query SearchQuery) ([]Deal, int64, error)
	FindByID(ctx context.Context, id uint64) (*Deal, error)
	FindBySlug(ctx context.Context, slug string) (*Deal, error)
	Upsert(ctx context.Context, deal Deal) (*Deal, error)
	UpdateStatus(ctx context.Context, id uint64, status string) (*Deal, error)
	Delete(ctx context.Context, id uint64) error
	Count(ctx context.Context) (int64, error)
}

type BrandRepository interface {
	List(ctx context.Context) ([]Brand, error)
	FindBySlug(ctx context.Context, slug string) (*Brand, error)
	FindByID(ctx context.Context, id uint64) (*Brand, error)
	Upsert(ctx context.Context, brand Brand) (*Brand, error)
	Update(ctx context.Context, brand Brand) (*Brand, error)
	Delete(ctx context.Context, id uint64) error
	ListSitemap(ctx context.Context) ([]SitemapEntry, error)
}

type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	FindByID(ctx context.Context, id uint64) (*Category, error)
	Upsert(ctx context.Context, category Category) (*Category, error)
	Update(ctx context.Context, category Category) (*Category, error)
	Delete(ctx context.Context, id uint64) error
	ListSitemap(ctx context.Context) ([]SitemapEntry, error)
}

type MerchantRepository interface {
	List(ctx context.Context) ([]Merchant, error)
	FindBySlug(ctx context.Context, slug string) (*Merchant, error)
	FindByID(ctx context.Context, id uint64) (*Merchant, error)
	Upsert(ctx context.Context, merchant Merchant) (*Merchant, error)
	Update(ctx context.Context, merchant Merchant) (*Merchant, error)
	Delete(ctx context.Context, id uint64) error
}

type CouponRepository interface {
	List(ctx context.Context, page, limit int) ([]Coupon, int64, error)
	ListByMerchants(ctx context.Context, merchantIDs []uint64) ([]Coupon, error)
	FindByID(ctx context.Context, id uint64) (*Coupon, error)
	Create(ctx context.Context, coupon Coupon) (*Coupon, error)
	Update(ctx context.Context, coupon Coupon) (*Coupon, error)
	Delete(ctx context.Context, id uint64) error
}

type SEORepository interface {
	Find(ctx context.Context, entityType string, entityID uint64) (*SEOMetadata, error)
	Upsert(ctx context.Context, meta SEOMetadata) (*SEOMetadata, error)
	List(ctx context.Context, page, limit int) ([]SEOMetadata, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type PriceAlertRepository interface {
	Create(ctx context.Context, alert PriceAlert) (*PriceAlert, error)
}

type AffiliateClickRepository interface {
	Create(ctx context.Context, click AffiliateClick) error
	SaveLink(ctx context.Context, offerID uint64, url, provider string) error
	CountToday(ctx context.Context) (int64, error)
}

type AdminRepository interface {
	Ensure(ctx context.Context, email, name, passwordHash string) (*Admin, error)
	FindByEmail(ctx context.Context, email string) (*Admin, error)
	FindByID(ctx context.Context, id uint64) (*Admin, error)
}

type SearchLogRepository interface {
	Create(ctx context.Context, query string, results int) error
}

type AuditRepository interface {
	Create(ctx context.Context, entry AuditEntry) error
}

type EventRepository interface {
	Create(ctx context.Context, event AnalyticsEvent) error
}
