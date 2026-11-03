package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/dealngon/backend/internal/domain"
)

type CatalogService struct {
	Products    domain.ProductRepository
	Variants    domain.VariantRepository
	Offers      domain.OfferRepository
	History     domain.PriceHistoryRepository
	Deals       domain.DealRepository
	Brands      domain.BrandRepository
	Categories  domain.CategoryRepository
	Merchants   domain.MerchantRepository
	Coupons     domain.CouponRepository
	SEO         domain.SEORepository
	Alerts      domain.PriceAlertRepository
	Searches    domain.SearchLogRepository
	Events      domain.EventRepository
	Cache       domain.Cache
	FrontendURL string
}

func (s *CatalogService) Search(ctx context.Context, query domain.SearchQuery) ([]domain.ProductCard, domain.PageMeta, error) {
	query.Page, query.Limit = domain.NormalizePage(query.Page, query.Limit)
	query.Sort = domain.NormalizeSort(query.Sort)
	items, total, err := s.Products.Search(ctx, query)
	if err != nil {
		return nil, domain.PageMeta{}, err
	}
	if items == nil {
		items = []domain.ProductCard{}
	}
	if strings.TrimSpace(query.Query) != "" && s.Searches != nil {
		_ = s.Searches.Create(ctx, strings.TrimSpace(query.Query), len(items))
	}
	return items, domain.PageMeta{Page: query.Page, Limit: query.Limit, Total: total}, nil
}

func (s *CatalogService) GetByID(ctx context.Context, id uint64) (*domain.ProductDetail, error) {
	product, err := s.Products.FindByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "PRODUCT_NOT_FOUND", "Không tìm thấy sản phẩm")
	}
	return s.detail(ctx, product)
}

func (s *CatalogService) GetBySlug(ctx context.Context, slug string) (*domain.ProductDetail, error) {
	key := "product:slug:" + slug
	if s.Cache != nil {
		if raw, ok, err := s.Cache.Get(ctx, key); err == nil && ok {
			var detail domain.ProductDetail
			if json.Unmarshal(raw, &detail) == nil {
				return &detail, nil
			}
		}
	}
	product, err := s.Products.FindBySlug(ctx, slug)
	if err != nil {
		return nil, notFound(err, "PRODUCT_NOT_FOUND", "Không tìm thấy sản phẩm")
	}
	detail, err := s.detail(ctx, product)
	if err != nil {
		return nil, err
	}
	if s.Cache != nil {
		if raw, err := json.Marshal(detail); err == nil {
			_ = s.Cache.Set(ctx, key, raw, time.Minute)
		}
	}
	return detail, nil
}

func (s *CatalogService) detail(ctx context.Context, product *domain.Product) (*domain.ProductDetail, error) {
	variants, err := s.Variants.ListByProduct(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	offers, err := s.Offers.ListByProduct(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	history, err := s.History.ListByProduct(ctx, product.ID)
	if err != nil {
		return nil, err
	}
	merchantIDs := make([]uint64, 0, len(offers))
	for i := range offers {
		offers[i].AffiliateURL = ""
		merchantIDs = append(merchantIDs, offers[i].MerchantID)
	}
	coupons, err := s.Coupons.ListByMerchants(ctx, merchantIDs)
	if err != nil {
		return nil, err
	}
	seo, err := s.SEO.Find(ctx, "product", product.ID)
	if errors.Is(err, domain.ErrNotFound) {
		seo = nil
	} else if err != nil {
		return nil, err
	}
	lowest, average, savings, last := domain.SummarizeOffers(offers)
	detail := &domain.ProductDetail{
		Product:       *product,
		Variants:      variants,
		Offers:        offers,
		LowestPrice:   lowest,
		AveragePrice:  average,
		Savings:       savings,
		LastCheckedAt: last,
		History:       history,
		Coupons:       coupons,
		SEO:           seo,
	}
	if deal, err := s.Deals.FindBySlug(ctx, product.Slug); err == nil {
		score := deal.Score
		detail.DealScore = &score
		detail.DealLabel = deal.Label
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if detail.Variants == nil {
		detail.Variants = []domain.Variant{}
	}
	if detail.Offers == nil {
		detail.Offers = []domain.Offer{}
	}
	if detail.History == nil {
		detail.History = []domain.PricePoint{}
	}
	if detail.Coupons == nil {
		detail.Coupons = []domain.Coupon{}
	}
	return detail, nil
}

func (s *CatalogService) ListOffers(ctx context.Context, productID uint64) ([]domain.Offer, error) {
	if _, err := s.Products.FindByID(ctx, productID); err != nil {
		return nil, notFound(err, "PRODUCT_NOT_FOUND", "Không tìm thấy sản phẩm")
	}
	offers, err := s.Offers.ListByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	for i := range offers {
		offers[i].AffiliateURL = ""
	}
	return offers, nil
}

func (s *CatalogService) PriceHistory(ctx context.Context, productID uint64) ([]domain.PricePoint, error) {
	if _, err := s.Products.FindByID(ctx, productID); err != nil {
		return nil, notFound(err, "PRODUCT_NOT_FOUND", "Không tìm thấy sản phẩm")
	}
	return s.History.ListByProduct(ctx, productID)
}

func (s *CatalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	const key = "categories:all"
	if s.Cache != nil {
		if raw, ok, err := s.Cache.Get(ctx, key); err == nil && ok {
			var items []domain.Category
			if json.Unmarshal(raw, &items) == nil {
				return items, nil
			}
		}
	}
	items, err := s.Categories.List(ctx)
	if err != nil {
		return nil, err
	}
	if s.Cache != nil {
		if raw, err := json.Marshal(items); err == nil {
			_ = s.Cache.Set(ctx, key, raw, 5*time.Minute)
		}
	}
	return items, nil
}

func (s *CatalogService) Category(ctx context.Context, slug string, page, limit int) (*domain.Category, []domain.ProductCard, domain.PageMeta, error) {
	category, err := s.Categories.FindBySlug(ctx, slug)
	if err != nil {
		return nil, nil, domain.PageMeta{}, notFound(err, "CATEGORY_NOT_FOUND", "Không tìm thấy danh mục")
	}
	items, meta, err := s.Search(ctx, domain.SearchQuery{CategorySlug: slug, Page: page, Limit: limit, Sort: "popular"})
	return category, items, meta, err
}

func (s *CatalogService) ListBrands(ctx context.Context) ([]domain.Brand, error) {
	return s.Brands.List(ctx)
}

func (s *CatalogService) Brand(ctx context.Context, slug string, page, limit int) (*domain.Brand, []domain.ProductCard, domain.PageMeta, error) {
	brand, err := s.Brands.FindBySlug(ctx, slug)
	if err != nil {
		return nil, nil, domain.PageMeta{}, notFound(err, "BRAND_NOT_FOUND", "Không tìm thấy thương hiệu")
	}
	items, meta, err := s.Search(ctx, domain.SearchQuery{BrandSlug: slug, Page: page, Limit: limit, Sort: "popular"})
	return brand, items, meta, err
}

func (s *CatalogService) ListMerchants(ctx context.Context) ([]domain.Merchant, error) {
	return s.Merchants.List(ctx)
}

func (s *CatalogService) ListDeals(ctx context.Context, query domain.SearchQuery) ([]domain.Deal, domain.PageMeta, error) {
	query.Page, query.Limit = domain.NormalizePage(query.Page, query.Limit)
	key := fmt.Sprintf("deals:list:%d:%d:%s", query.Page, query.Limit, query.Sort)
	if s.Cache != nil && query.Query == "" {
		if raw, ok, err := s.Cache.Get(ctx, key); err == nil && ok {
			var cached struct {
				Items []domain.Deal   `json:"items"`
				Meta  domain.PageMeta `json:"meta"`
			}
			if json.Unmarshal(raw, &cached) == nil {
				return cached.Items, cached.Meta, nil
			}
		}
	}
	items, total, err := s.Deals.List(ctx, query)
	if err != nil {
		return nil, domain.PageMeta{}, err
	}
	if items == nil {
		items = []domain.Deal{}
	}
	meta := domain.PageMeta{Page: query.Page, Limit: query.Limit, Total: total}
	if s.Cache != nil {
		if raw, err := json.Marshal(struct {
			Items []domain.Deal   `json:"items"`
			Meta  domain.PageMeta `json:"meta"`
		}{items, meta}); err == nil {
			_ = s.Cache.Set(ctx, key, raw, 30*time.Second)
		}
	}
	return items, meta, nil
}

func (s *CatalogService) DealByID(ctx context.Context, id uint64) (*domain.Deal, error) {
	deal, err := s.Deals.FindByID(ctx, id)
	if err != nil {
		return nil, notFound(err, "DEAL_NOT_FOUND", "Không tìm thấy deal")
	}
	return deal, nil
}

func (s *CatalogService) DealBySlug(ctx context.Context, slug string) (*domain.Deal, error) {
	deal, err := s.Deals.FindBySlug(ctx, slug)
	if err != nil {
		return nil, notFound(err, "DEAL_NOT_FOUND", "Không tìm thấy deal")
	}
	return deal, nil
}

func (s *CatalogService) CreateAlert(ctx context.Context, alert domain.PriceAlert) (*domain.PriceAlert, error) {
	if _, err := mail.ParseAddress(alert.Email); err != nil {
		return nil, domain.NewError("INVALID_EMAIL", "Email không hợp lệ", domain.ErrInvalid)
	}
	if alert.TargetPrice <= 0 {
		return nil, domain.NewError("INVALID_PRICE", "Giá mục tiêu phải lớn hơn 0", domain.ErrInvalid)
	}
	if _, err := s.Products.FindByID(ctx, alert.ProductID); err != nil {
		return nil, notFound(err, "PRODUCT_NOT_FOUND", "Không tìm thấy sản phẩm")
	}
	return s.Alerts.Create(ctx, alert)
}

func (s *CatalogService) TrackEvent(ctx context.Context, event domain.AnalyticsEvent) error {
	if strings.TrimSpace(event.Name) == "" {
		return domain.NewError("INVALID_EVENT", "Thiếu tên sự kiện", domain.ErrInvalid)
	}
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	return s.Events.Create(ctx, event)
}

func (s *CatalogService) Sitemap(ctx context.Context) (string, error) {
	base := strings.TrimRight(s.FrontendURL, "/")
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	builder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	write := func(path string, updated time.Time) {
		builder.WriteString("<url><loc>")
		builder.WriteString(xmlEscape(base + path))
		builder.WriteString("</loc>")
		if !updated.IsZero() {
			builder.WriteString("<lastmod>")
			builder.WriteString(updated.UTC().Format("2006-01-02"))
			builder.WriteString("</lastmod>")
		}
		builder.WriteString("</url>")
	}
	for _, path := range []string{"/", "/tim-kiem", "/deal", "/gioi-thieu", "/lien-he", "/bao-mat", "/dieu-khoan", "/chinh-sach-affiliate"} {
		write(path, time.Time{})
	}
	products, err := s.Products.ListSitemap(ctx)
	if err != nil {
		return "", err
	}
	deals, _, err := s.Deals.List(ctx, domain.SearchQuery{Page: 1, Limit: 100})
	if err != nil {
		return "", err
	}
	brands, err := s.Brands.ListSitemap(ctx)
	if err != nil {
		return "", err
	}
	categories, err := s.Categories.ListSitemap(ctx)
	if err != nil {
		return "", err
	}
	for _, entry := range products {
		write(entry.Path, entry.UpdatedAt)
	}
	for _, deal := range deals {
		write("/deal/"+deal.Slug, deal.DetectedAt)
	}
	for _, entry := range brands {
		write(entry.Path, entry.UpdatedAt)
	}
	for _, entry := range categories {
		write(entry.Path, entry.UpdatedAt)
	}
	builder.WriteString("</urlset>")
	return builder.String(), nil
}

func (s *CatalogService) Invalidate(ctx context.Context) {
	if s.Cache == nil {
		return
	}
	_ = s.Cache.DeletePrefix(ctx, "product:slug:")
	_ = s.Cache.DeletePrefix(ctx, "deals:list:")
	_ = s.Cache.DeletePrefix(ctx, "categories:")
}

func notFound(err error, code, message string) error {
	if errors.Is(err, domain.ErrNotFound) {
		return domain.NewError(code, message, domain.ErrNotFound)
	}
	return err
}

func xmlEscape(value string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(value)
}
