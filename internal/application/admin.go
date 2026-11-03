package application

import (
	"context"
	"errors"
	"strings"

	"github.com/dealngon/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type TokenIssuer interface {
	Issue(adminID uint64, email string) (string, error)
	Parse(token string) (uint64, error)
}

type AdminService struct {
	Admins     domain.AdminRepository
	Products   domain.ProductRepository
	Variants   domain.VariantRepository
	Offers     domain.OfferRepository
	Brands     domain.BrandRepository
	Categories domain.CategoryRepository
	Merchants  domain.MerchantRepository
	Deals      domain.DealRepository
	Coupons    domain.CouponRepository
	SEO        domain.SEORepository
	Clicks     domain.AffiliateClickRepository
	Audits     domain.AuditRepository
	Tokens     TokenIssuer
	Sync       *SyncService
	Catalog    *CatalogService
}

type LoginResult struct {
	Token string       `json:"token"`
	Admin domain.Admin `json:"admin"`
}

func (s *AdminService) Ensure(ctx context.Context, email, name, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.Admins.Ensure(ctx, strings.ToLower(strings.TrimSpace(email)), name, string(hash))
	return err
}

func (s *AdminService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	admin, err := s.Admins.FindByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.NewError("INVALID_CREDENTIALS", "Email hoặc mật khẩu không đúng", domain.ErrUnauthorized)
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return nil, domain.NewError("INVALID_CREDENTIALS", "Email hoặc mật khẩu không đúng", domain.ErrUnauthorized)
	}
	token, err := s.Tokens.Issue(admin.ID, admin.Email)
	if err != nil {
		return nil, err
	}
	admin.PasswordHash = ""
	return &LoginResult{Token: token, Admin: *admin}, nil
}

func (s *AdminService) Dashboard(ctx context.Context) (*domain.DashboardStats, error) {
	products, err := s.Products.Count(ctx)
	if err != nil {
		return nil, err
	}
	offers, err := s.Offers.Count(ctx)
	if err != nil {
		return nil, err
	}
	deals, err := s.Deals.Count(ctx)
	if err != nil {
		return nil, err
	}
	clicks, err := s.Clicks.CountToday(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.DashboardStats{Products: products, Offers: offers, Deals: deals, ClicksToday: clicks}, nil
}

func (s *AdminService) audit(ctx context.Context, adminID uint64, action, entity string, id uint64) {
	if s.Audits == nil || adminID == 0 {
		return
	}
	_ = s.Audits.Create(ctx, domain.AuditEntry{AdminID: adminID, Action: action, EntityType: entity, EntityID: id})
}

func (s *AdminService) invalidate(ctx context.Context) {
	if s.Catalog != nil {
		s.Catalog.Invalidate(ctx)
	}
}

func normalizeSlug(name, slug string) (string, error) {
	if strings.TrimSpace(name) == "" && strings.TrimSpace(slug) == "" {
		return "", domain.NewError("INVALID_NAME", "Thiếu tên", domain.ErrInvalid)
	}
	if strings.TrimSpace(slug) == "" {
		slug = domain.Slugify(name)
	} else {
		slug = domain.Slugify(slug)
	}
	if slug == "" {
		return "", domain.NewError("INVALID_SLUG", "Slug không hợp lệ", domain.ErrInvalid)
	}
	return slug, nil
}

func (s *AdminService) SaveBrand(ctx context.Context, adminID uint64, brand domain.Brand) (*domain.Brand, error) {
	slug, err := normalizeSlug(brand.Name, brand.Slug)
	if err != nil {
		return nil, err
	}
	brand.Slug = slug
	var saved *domain.Brand
	action := "update"
	if brand.ID == 0 {
		saved, err = s.Brands.Upsert(ctx, brand)
		action = "create"
	} else {
		saved, err = s.Brands.Update(ctx, brand)
	}
	if err != nil {
		return nil, err
	}
	s.audit(ctx, adminID, action, "brand", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteBrand(ctx context.Context, adminID, id uint64) error {
	err := s.Brands.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "brand", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) SaveCategory(ctx context.Context, adminID uint64, category domain.Category) (*domain.Category, error) {
	slug, err := normalizeSlug(category.Name, category.Slug)
	if err != nil {
		return nil, err
	}
	category.Slug = slug
	var saved *domain.Category
	action := "update"
	if category.ID == 0 {
		saved, err = s.Categories.Upsert(ctx, category)
		action = "create"
	} else {
		saved, err = s.Categories.Update(ctx, category)
	}
	if err != nil {
		return nil, err
	}
	s.audit(ctx, adminID, action, "category", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteCategory(ctx context.Context, adminID, id uint64) error {
	err := s.Categories.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "category", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) SaveMerchant(ctx context.Context, adminID uint64, merchant domain.Merchant) (*domain.Merchant, error) {
	slug, err := normalizeSlug(merchant.Name, merchant.Slug)
	if err != nil {
		return nil, err
	}
	merchant.Slug = slug
	if merchant.Status == "" {
		merchant.Status = domain.StatusActive
	}
	if merchant.QualityScore < 0 || merchant.QualityScore > 5 {
		return nil, domain.NewError("INVALID_QUALITY", "Điểm chất lượng merchant từ 0 đến 5", domain.ErrInvalid)
	}
	var saved *domain.Merchant
	action := "update"
	if merchant.ID == 0 {
		saved, err = s.Merchants.Upsert(ctx, merchant)
		action = "create"
	} else {
		saved, err = s.Merchants.Update(ctx, merchant)
	}
	if err != nil {
		return nil, err
	}
	s.audit(ctx, adminID, action, "merchant", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteMerchant(ctx context.Context, adminID, id uint64) error {
	err := s.Merchants.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "merchant", id)
		s.invalidate(ctx)
	}
	return err
}

type ProductWrite struct {
	Product domain.Product  `json:"product"`
	Variant *domain.Variant `json:"variant"`
}

func (s *AdminService) ListProducts(ctx context.Context, page, limit int) ([]domain.Product, domain.PageMeta, error) {
	page, limit = domain.NormalizePage(page, limit)
	items, total, err := s.Products.List(ctx, page, limit)
	if items == nil {
		items = []domain.Product{}
	}
	return items, domain.PageMeta{Page: page, Limit: limit, Total: total}, err
}

func (s *AdminService) SaveProduct(ctx context.Context, adminID uint64, input ProductWrite) (*domain.Product, error) {
	slug, err := normalizeSlug(input.Product.Name, input.Product.Slug)
	if err != nil {
		return nil, err
	}
	if input.Product.BrandID == 0 || input.Product.CategoryID == 0 {
		return nil, domain.NewError("INVALID_PRODUCT", "Thiếu thương hiệu hoặc danh mục", domain.ErrInvalid)
	}
	input.Product.Slug = slug
	if input.Product.Status == "" {
		input.Product.Status = domain.StatusActive
	}
	if input.Product.Specifications == nil {
		input.Product.Specifications = map[string]string{}
	}
	var saved *domain.Product
	if input.Product.ID == 0 {
		saved, err = s.Products.Create(ctx, input.Product)
	} else {
		saved, err = s.Products.Update(ctx, input.Product)
	}
	if err != nil {
		return nil, err
	}
	if input.Variant != nil && strings.TrimSpace(input.Variant.SKU) != "" {
		input.Variant.ProductID = saved.ID
		if _, err := s.Variants.Upsert(ctx, *input.Variant); err != nil {
			return nil, err
		}
	}
	action := "update"
	if input.Product.ID == 0 {
		action = "create"
	}
	s.audit(ctx, adminID, action, "product", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteProduct(ctx context.Context, adminID, id uint64) error {
	err := s.Products.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "product", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) ListOffers(ctx context.Context, page, limit int) ([]domain.Offer, domain.PageMeta, error) {
	page, limit = domain.NormalizePage(page, limit)
	items, total, err := s.Offers.List(ctx, page, limit)
	for i := range items {
		items[i].AffiliateURL = ""
	}
	return items, domain.PageMeta{Page: page, Limit: limit, Total: total}, err
}

func (s *AdminService) SaveOffer(ctx context.Context, adminID uint64, offer domain.Offer) (*domain.Offer, error) {
	if offer.ProductVariantID == 0 || offer.MerchantID == 0 || offer.Price <= 0 {
		return nil, domain.NewError("INVALID_OFFER", "Offer cần biến thể, nhà bán và giá lớn hơn 0", domain.ErrInvalid)
	}
	if offer.OriginalPrice != nil && *offer.OriginalPrice < offer.Price {
		return nil, domain.NewError("INVALID_PRICE", "Giá gốc không được thấp hơn giá bán", domain.ErrInvalid)
	}
	if offer.Status == "" {
		offer.Status = domain.StatusActive
	}
	var saved *domain.Offer
	var err error
	if offer.ID == 0 {
		saved, err = s.Offers.Upsert(ctx, offer)
	} else {
		saved, err = s.Offers.Update(ctx, offer)
	}
	if err != nil {
		return nil, err
	}
	if saved.ProductID != 0 && s.Sync != nil {
		_ = s.Sync.RecomputeProduct(ctx, saved.ProductID)
	}
	action := "update"
	if offer.ID == 0 {
		action = "create"
	}
	s.audit(ctx, adminID, action, "offer", saved.ID)
	s.invalidate(ctx)
	saved.AffiliateURL = ""
	return saved, nil
}

func (s *AdminService) DeleteOffer(ctx context.Context, adminID, id uint64) error {
	err := s.Offers.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "offer", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) ListDeals(ctx context.Context, page, limit int) ([]domain.Deal, domain.PageMeta, error) {
	page, limit = domain.NormalizePage(page, limit)
	items, total, err := s.Deals.List(ctx, domain.SearchQuery{Page: page, Limit: limit, Sort: "deal"})
	return items, domain.PageMeta{Page: page, Limit: limit, Total: total}, err
}

func (s *AdminService) UpdateDealStatus(ctx context.Context, adminID, id uint64, status string) (*domain.Deal, error) {
	if status != domain.StatusActive && status != domain.StatusInactive {
		return nil, domain.NewError("INVALID_STATUS", "Trạng thái không hợp lệ", domain.ErrInvalid)
	}
	deal, err := s.Deals.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, adminID, "update", "deal", id)
	s.invalidate(ctx)
	return deal, nil
}

func (s *AdminService) DeleteDeal(ctx context.Context, adminID, id uint64) error {
	err := s.Deals.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "deal", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) Recompute(ctx context.Context, adminID uint64) error {
	if s.Sync == nil {
		return domain.NewError("SYNC_UNAVAILABLE", "Không thể đồng bộ", domain.ErrInvalid)
	}
	if err := s.Sync.Run(ctx); err != nil {
		return err
	}
	s.audit(ctx, adminID, "sync", "catalog", 0)
	return nil
}

func (s *AdminService) ListCoupons(ctx context.Context, page, limit int) ([]domain.Coupon, domain.PageMeta, error) {
	page, limit = domain.NormalizePage(page, limit)
	items, total, err := s.Coupons.List(ctx, page, limit)
	return items, domain.PageMeta{Page: page, Limit: limit, Total: total}, err
}

func (s *AdminService) SaveCoupon(ctx context.Context, adminID uint64, coupon domain.Coupon) (*domain.Coupon, error) {
	if coupon.MerchantID == 0 || strings.TrimSpace(coupon.Code) == "" {
		return nil, domain.NewError("INVALID_COUPON", "Coupon cần nhà bán và mã", domain.ErrInvalid)
	}
	if coupon.Status == "" {
		coupon.Status = domain.StatusActive
	}
	var saved *domain.Coupon
	var err error
	if coupon.ID == 0 {
		saved, err = s.Coupons.Create(ctx, coupon)
	} else {
		saved, err = s.Coupons.Update(ctx, coupon)
	}
	if err != nil {
		return nil, err
	}
	action := "update"
	if coupon.ID == 0 {
		action = "create"
	}
	s.audit(ctx, adminID, action, "coupon", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteCoupon(ctx context.Context, adminID, id uint64) error {
	err := s.Coupons.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "coupon", id)
		s.invalidate(ctx)
	}
	return err
}

func (s *AdminService) ListSEO(ctx context.Context, page, limit int) ([]domain.SEOMetadata, domain.PageMeta, error) {
	page, limit = domain.NormalizePage(page, limit)
	items, total, err := s.SEO.List(ctx, page, limit)
	return items, domain.PageMeta{Page: page, Limit: limit, Total: total}, err
}

func (s *AdminService) SaveSEO(ctx context.Context, adminID uint64, meta domain.SEOMetadata) (*domain.SEOMetadata, error) {
	if strings.TrimSpace(meta.EntityType) == "" || meta.EntityID == 0 {
		return nil, domain.NewError("INVALID_SEO", "Thiếu loại và mã thực thể", domain.ErrInvalid)
	}
	saved, err := s.SEO.Upsert(ctx, meta)
	if err != nil {
		return nil, err
	}
	s.audit(ctx, adminID, "upsert", "seo", saved.ID)
	s.invalidate(ctx)
	return saved, nil
}

func (s *AdminService) DeleteSEO(ctx context.Context, adminID, id uint64) error {
	err := s.SEO.Delete(ctx, id)
	if err == nil {
		s.audit(ctx, adminID, "delete", "seo", id)
		s.invalidate(ctx)
	}
	return err
}
