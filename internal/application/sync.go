package application

import (
	"context"
	"errors"
	"time"

	"github.com/dealngon/backend/internal/domain"
)

type SyncService struct {
	Providers      []domain.MerchantProvider
	Matcher        domain.ProductMatcher
	Brands         domain.BrandRepository
	Categories     domain.CategoryRepository
	Merchants      domain.MerchantRepository
	Products       domain.ProductRepository
	Variants       domain.VariantRepository
	Offers         domain.OfferRepository
	History        domain.PriceHistoryRepository
	Deals          domain.DealRepository
	Cache          domain.Cache
	SeedMerchants  []domain.Merchant
	CategoryLabels map[string]string
	Now            func() time.Time
}

func (s *SyncService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now().UTC()
}

func (s *SyncService) Run(ctx context.Context) error {
	if err := s.ensureMerchants(ctx); err != nil {
		return err
	}
	for _, provider := range s.Providers {
		products, err := provider.FetchProducts(ctx)
		if err != nil {
			return err
		}
		offers, err := provider.FetchOffers(ctx)
		if err != nil {
			return err
		}
		bySKU := map[string]domain.ExternalOffer{}
		for _, offer := range offers {
			bySKU[offer.SKU] = offer
		}
		for _, item := range products {
			if offer, ok := bySKU[item.SKU]; ok && offer.Price > 0 {
				item.Price = offer.Price
				item.OriginalPrice = offer.OriginalPrice
				if offer.ProductURL != "" {
					item.ProductURL = offer.ProductURL
				}
				if offer.SellerName != "" {
					item.SellerName = offer.SellerName
				}
				item.Stock = offer.Stock
				item.Rating = offer.Rating
			}
			if err := s.upsertItem(ctx, item); err != nil {
				return err
			}
		}
	}
	if err := s.RecomputeAll(ctx); err != nil {
		return err
	}
	if s.Cache != nil {
		_ = s.Cache.DeletePrefix(ctx, "product:slug:")
		_ = s.Cache.DeletePrefix(ctx, "deals:list:")
		_ = s.Cache.DeletePrefix(ctx, "categories:")
	}
	return nil
}

func (s *SyncService) ensureMerchants(ctx context.Context) error {
	for _, info := range s.SeedMerchants {
		merchant := info
		if merchant.Status == "" {
			merchant.Status = domain.StatusActive
		}
		merchant.AffiliateEnabled = true
		if _, err := s.Merchants.Upsert(ctx, merchant); err != nil {
			return err
		}
	}
	for slug, name := range s.CategoryLabels {
		if _, err := s.Categories.Upsert(ctx, domain.Category{Name: name, Slug: slug}); err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncService) upsertItem(ctx context.Context, item domain.ExternalProduct) error {
	if item.Price <= 0 || item.SKU == "" {
		return domain.NewError("INVALID_OFFER", "Dữ liệu offer không hợp lệ", domain.ErrInvalid)
	}
	merchantRow, err := s.Merchants.FindBySlug(ctx, item.MerchantSlug)
	if err != nil {
		return err
	}
	variant, err := s.Matcher.Match(ctx, item)
	if errors.Is(err, domain.ErrNotFound) {
		variant, err = s.createCatalog(ctx, item)
	}
	if err != nil {
		return err
	}
	checked := s.now()
	offer, err := s.Offers.Upsert(ctx, domain.Offer{
		ProductVariantID: variant.ID,
		MerchantID:       merchantRow.ID,
		Title:            item.Name,
		ProductURL:       item.ProductURL,
		Price:            item.Price,
		OriginalPrice:    item.OriginalPrice,
		Stock:            item.Stock,
		Rating:           item.Rating,
		SellerName:       item.SellerName,
		Status:           domain.StatusActive,
		LastCheckedAt:    &checked,
	})
	if err != nil {
		return err
	}
	return s.writeHistory(ctx, offer.ID, item.SKU, item.Price)
}

func (s *SyncService) createCatalog(ctx context.Context, item domain.ExternalProduct) (*domain.Variant, error) {
	brand, err := s.Brands.Upsert(ctx, domain.Brand{Name: item.Brand, Slug: domain.Slugify(item.Brand)})
	if err != nil {
		return nil, err
	}
	categoryName := item.Category
	if name, ok := s.CategoryLabels[item.Category]; ok {
		categoryName = name
	}
	category, err := s.Categories.Upsert(ctx, domain.Category{Name: categoryName, Slug: domain.Slugify(item.Category)})
	if err != nil {
		return nil, err
	}
	slug := domain.Slugify(item.Name)
	product, err := s.Products.FindBySlug(ctx, slug)
	if errors.Is(err, domain.ErrNotFound) {
		specs := item.Specifications
		if specs == nil {
			specs = map[string]string{}
		}
		product, err = s.Products.Create(ctx, domain.Product{
			BrandID:        brand.ID,
			CategoryID:     category.ID,
			Name:           item.Name,
			Slug:           slug,
			Description:    item.Description,
			Image:          item.Image,
			Specifications: specs,
			Status:         domain.StatusActive,
		})
		if errors.Is(err, domain.ErrConflict) {
			product, err = s.Products.FindBySlug(ctx, slug)
		}
	}
	if err != nil {
		return nil, err
	}
	return s.Variants.Upsert(ctx, domain.Variant{
		ProductID:   product.ID,
		SKU:         item.SKU,
		Model:       item.Model,
		Storage:     item.Storage,
		RAM:         item.RAM,
		Color:       item.Color,
		Barcode:     item.Barcode,
		VariantName: item.VariantName,
	})
}

func (s *SyncService) writeHistory(ctx context.Context, offerID uint64, sku string, price int64) error {
	has, err := s.History.HasHistory(ctx, offerID)
	if err != nil {
		return err
	}
	now := s.now()
	if !has {
		points := make([]domain.PricePoint, 0, 31)
		for days := 30; days >= 0; days-- {
			points = append(points, domain.PricePoint{
				OfferID:    offerID,
				Price:      domain.HistoricalPrice(price, days, domain.AllowPriceDip(sku)),
				RecordedAt: now.AddDate(0, 0, -days),
			})
		}
		return s.History.Backfill(ctx, offerID, points)
	}
	return s.History.AppendIfChanged(ctx, offerID, price, now)
}

func (s *SyncService) RecomputeAll(ctx context.Context) error {
	ids, err := s.Products.ListIDs(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := s.RecomputeProduct(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *SyncService) RecomputeProduct(ctx context.Context, productID uint64) error {
	product, err := s.Products.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	offers, err := s.Offers.ListByProduct(ctx, productID)
	if err != nil {
		return err
	}
	best := domain.BestOffer(offers)
	if best == nil {
		return nil
	}
	points, err := s.History.ListByOffer(ctx, best.ID)
	if err != nil {
		return err
	}
	avg30, min30, historical := domain.SummarizeHistory(points, s.now())
	result := domain.CalculateDealScore(domain.DealScoreInput{
		CurrentPrice:    best.Price,
		OriginalPrice:   best.OriginalPrice,
		Avg30:           avg30,
		Min30:           min30,
		HistoricalMin:   historical,
		MerchantQuality: best.QualityScore,
	})
	lowest, average, _, _ := domain.SummarizeOffers(offers)
	low := best.Price
	avg := best.Price
	if lowest != nil {
		low = *lowest
	}
	if average != nil {
		avg = *average
	}
	_, err = s.Deals.Upsert(ctx, domain.Deal{
		ProductID:       product.ID,
		OfferID:         best.ID,
		Slug:            product.Slug,
		Score:           result.Score,
		Label:           result.Label,
		DiscountAmount:  result.DiscountAmount,
		DiscountPercent: result.DiscountPercent,
		LowestPrice:     low,
		AveragePrice:    avg,
		Status:          domain.StatusActive,
		DetectedAt:      s.now(),
	})
	return err
}
