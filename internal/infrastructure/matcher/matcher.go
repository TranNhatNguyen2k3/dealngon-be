package matcher

import (
	"context"
	"errors"

	"github.com/dealngon/backend/internal/domain"
)

type Service struct {
	Variants domain.VariantRepository
}

func New(variants domain.VariantRepository) *Service {
	return &Service{Variants: variants}
}

func (s *Service) Match(ctx context.Context, item domain.ExternalProduct) (*domain.Variant, error) {
	if item.Barcode != "" {
		variant, err := s.Variants.FindByBarcode(ctx, item.Barcode)
		if err == nil {
			return variant, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
	}
	if item.SKU != "" {
		variant, err := s.Variants.FindBySKU(ctx, item.SKU)
		if err == nil {
			return variant, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
	}
	brand := domain.Slugify(item.Brand)
	if brand != "" && item.Model != "" {
		variant, err := s.Variants.FindByBrandModel(ctx, brand, item.Model)
		if err == nil {
			return variant, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
		variant, err = s.Variants.FindByAttributes(ctx, brand, item.Model, item.Storage, item.RAM, item.Color)
		if err == nil {
			return variant, nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return nil, err
		}
	}
	variant, err := s.Variants.FindByNormalizedName(ctx, item.Name)
	if err == nil {
		return variant, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	return nil, domain.ErrNotFound
}
