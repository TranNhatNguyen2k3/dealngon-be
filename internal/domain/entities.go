package domain

import "time"

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Brand struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Logo      string    `json:"logo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	ParentID  *uint64   `json:"parent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID             uint64            `json:"id"`
	BrandID        uint64            `json:"brand_id"`
	CategoryID     uint64            `json:"category_id"`
	Name           string            `json:"name"`
	Slug           string            `json:"slug"`
	Description    string            `json:"description"`
	Image          string            `json:"image"`
	Specifications map[string]string `json:"specifications"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	BrandName      string            `json:"brand_name,omitempty"`
	BrandSlug      string            `json:"brand_slug,omitempty"`
	CategoryName   string            `json:"category_name,omitempty"`
	CategorySlug   string            `json:"category_slug,omitempty"`
}

type Variant struct {
	ID          uint64    `json:"id"`
	ProductID   uint64    `json:"product_id"`
	SKU         string    `json:"sku"`
	Model       string    `json:"model"`
	Storage     string    `json:"storage"`
	RAM         string    `json:"ram"`
	Color       string    `json:"color"`
	Barcode     string    `json:"barcode"`
	VariantName string    `json:"variant_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Merchant struct {
	ID               uint64    `json:"id"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	Logo             string    `json:"logo"`
	Website          string    `json:"website"`
	AffiliateEnabled bool      `json:"affiliate_enabled"`
	Status           string    `json:"status"`
	QualityScore     int       `json:"quality_score"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Offer struct {
	ID               uint64     `json:"id"`
	ProductVariantID uint64     `json:"product_variant_id"`
	ProductID        uint64     `json:"product_id,omitempty"`
	MerchantID       uint64     `json:"merchant_id"`
	Title            string     `json:"title"`
	ProductURL       string     `json:"product_url"`
	AffiliateURL     string     `json:"affiliate_url,omitempty"`
	Price            int64      `json:"price"`
	OriginalPrice    *int64     `json:"original_price"`
	Stock            *int       `json:"stock"`
	Rating           *float64   `json:"rating"`
	SellerName       string     `json:"seller_name"`
	Status           string     `json:"status"`
	LastCheckedAt    *time.Time `json:"last_checked_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	MerchantName     string     `json:"merchant_name,omitempty"`
	MerchantSlug     string     `json:"merchant_slug,omitempty"`
	MerchantLogo     string     `json:"merchant_logo,omitempty"`
	VariantName      string     `json:"variant_name,omitempty"`
	QualityScore     int        `json:"-"`
}

type PricePoint struct {
	OfferID      uint64    `json:"offer_id"`
	MerchantName string    `json:"merchant_name,omitempty"`
	Price        int64     `json:"price"`
	RecordedAt   time.Time `json:"recorded_at"`
}

type Deal struct {
	ID              uint64    `json:"id"`
	ProductID       uint64    `json:"product_id"`
	OfferID         uint64    `json:"offer_id"`
	Slug            string    `json:"slug"`
	Score           int       `json:"score"`
	Label           string    `json:"label"`
	DiscountAmount  *int64    `json:"discount_amount"`
	DiscountPercent *float64  `json:"discount_percent"`
	LowestPrice     int64     `json:"lowest_price"`
	AveragePrice    int64     `json:"average_price"`
	Status          string    `json:"status"`
	DetectedAt      time.Time `json:"detected_at"`
	ProductName     string    `json:"product_name,omitempty"`
	ProductSlug     string    `json:"product_slug,omitempty"`
	ProductImage    string    `json:"product_image,omitempty"`
	BrandName       string    `json:"brand_name,omitempty"`
	MerchantName    string    `json:"merchant_name,omitempty"`
	CategorySlug    string    `json:"category_slug,omitempty"`
}

type Coupon struct {
	ID              uint64     `json:"id"`
	MerchantID      uint64     `json:"merchant_id"`
	Code            string     `json:"code"`
	Description     string     `json:"description"`
	DiscountPercent *float64   `json:"discount_percent"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	Status          string     `json:"status"`
	MerchantName    string     `json:"merchant_name,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type SEOMetadata struct {
	ID          uint64    `json:"id"`
	EntityType  string    `json:"entity_type"`
	EntityID    uint64    `json:"entity_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Canonical   string    `json:"canonical"`
	OGImage     string    `json:"og_image"`
	Robots      string    `json:"robots"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PriceAlert struct {
	ID          uint64    `json:"id"`
	ProductID   uint64    `json:"product_id"`
	Email       string    `json:"email"`
	TargetPrice int64     `json:"target_price"`
	CreatedAt   time.Time `json:"created_at"`
}

type AffiliateClick struct {
	ID          uint64    `json:"id"`
	OfferID     uint64    `json:"offer_id"`
	ProductID   uint64    `json:"product_id"`
	MerchantID  uint64    `json:"merchant_id"`
	SessionID   string    `json:"session_id"`
	UserID      *uint64   `json:"user_id"`
	IPHash      string    `json:"ip_hash"`
	UserAgent   string    `json:"user_agent"`
	Referer     string    `json:"referer"`
	UTMSource   string    `json:"utm_source"`
	UTMMedium   string    `json:"utm_medium"`
	UTMCampaign string    `json:"utm_campaign"`
	CreatedAt   time.Time `json:"created_at"`
}

type Admin struct {
	ID           uint64    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProductCard struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Image           string     `json:"image"`
	BrandName       string     `json:"brand_name"`
	BrandSlug       string     `json:"brand_slug"`
	CategoryName    string     `json:"category_name"`
	CategorySlug    string     `json:"category_slug"`
	LowestPrice     *int64     `json:"lowest_price"`
	OriginalPrice   *int64     `json:"original_price"`
	DiscountPercent *float64   `json:"discount_percent"`
	DealScore       *int       `json:"deal_score"`
	DealLabel       string     `json:"deal_label"`
	OfferCount      int        `json:"offer_count"`
	LastCheckedAt   *time.Time `json:"last_checked_at"`
}

type ProductDetail struct {
	Product       Product      `json:"product"`
	Variants      []Variant    `json:"variants"`
	Offers        []Offer      `json:"offers"`
	LowestPrice   *int64       `json:"lowest_price"`
	AveragePrice  *int64       `json:"average_price"`
	Savings       *int64       `json:"savings"`
	DealScore     *int         `json:"deal_score"`
	DealLabel     string       `json:"deal_label"`
	LastCheckedAt *time.Time   `json:"last_checked_at"`
	History       []PricePoint `json:"price_history"`
	Coupons       []Coupon     `json:"coupons"`
	SEO           *SEOMetadata `json:"seo"`
}

type SearchQuery struct {
	Query        string
	CategorySlug string
	BrandSlug    string
	MerchantSlug string
	MinPrice     *int64
	MaxPrice     *int64
	MinDiscount  *float64
	Sort         string
	Page         int
	Limit        int
}

type PageMeta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

type DashboardStats struct {
	Products    int64 `json:"products"`
	Offers      int64 `json:"offers"`
	Deals       int64 `json:"deals"`
	ClicksToday int64 `json:"clicks_today"`
}

type SitemapEntry struct {
	Path      string
	UpdatedAt time.Time
}

type AnalyticsEvent struct {
	Name      string
	SessionID string
	Payload   map[string]any
}

type AuditEntry struct {
	AdminID    uint64
	Action     string
	EntityType string
	EntityID   uint64
}
