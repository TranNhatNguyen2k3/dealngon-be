package postgres

import "github.com/dealngon/backend/internal/domain"

var (
	_ domain.ProductRepository      = (*ProductRepo)(nil)
	_ domain.VariantRepository      = (*VariantRepo)(nil)
	_ domain.OfferRepository        = (*OfferRepo)(nil)
	_ domain.PriceHistoryRepository = (*HistoryRepo)(nil)
	_ domain.DealRepository         = (*DealRepo)(nil)
	_ domain.BrandRepository        = (*BrandRepo)(nil)
	_ domain.CategoryRepository     = (*CategoryRepo)(nil)
	_ domain.MerchantRepository     = (*MerchantRepo)(nil)
	_ domain.CouponRepository       = (*CouponRepo)(nil)
	_ domain.SEORepository          = (*SEORepo)(nil)
	_ domain.PriceAlertRepository   = (*AlertRepo)(nil)
	_ domain.AffiliateClickRepository = (*ClickRepo)(nil)
	_ domain.AdminRepository        = (*AdminRepo)(nil)
	_ domain.SearchLogRepository    = (*SearchLogRepo)(nil)
	_ domain.AuditRepository        = (*AuditRepo)(nil)
	_ domain.EventRepository        = (*EventRepo)(nil)
)
