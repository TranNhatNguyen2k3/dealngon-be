package application

import (
	"context"
	"errors"

	"github.com/dealngon/backend/internal/domain"
)

type AffiliateService struct {
	Offers   domain.OfferRepository
	Clicks   domain.AffiliateClickRepository
	Provider domain.AffiliateProvider
}

type ClickInput struct {
	OfferID     uint64
	SessionID   string
	UserID      *uint64
	IPHash      string
	UserAgent   string
	Referer     string
	UTMSource   string
	UTMMedium   string
	UTMCampaign string
}

func (s *AffiliateService) Redirect(ctx context.Context, input ClickInput) (string, error) {
	offer, err := s.Offers.FindByID(ctx, input.OfferID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", domain.NewError("OFFER_NOT_FOUND", "Không tìm thấy offer", domain.ErrNotFound)
		}
		return "", err
	}
	if offer.Status != domain.StatusActive {
		return "", domain.NewError("OFFER_INACTIVE", "Offer không còn hiệu lực", domain.ErrNotFound)
	}
	link, err := s.Provider.GenerateLink(ctx, *offer)
	if err != nil {
		return "", err
	}
	providerName := "mock"
	if s.Provider != nil && s.Provider.Name() != "" {
		providerName = s.Provider.Name()
	}
	if err := s.Clicks.SaveLink(ctx, offer.ID, link, providerName); err != nil {
		return "", err
	}
	if err := s.Offers.SetAffiliateURL(ctx, offer.ID, link); err != nil {
		return "", err
	}
	if err := s.Clicks.Create(ctx, domain.AffiliateClick{
		OfferID:     offer.ID,
		ProductID:   offer.ProductID,
		MerchantID:  offer.MerchantID,
		SessionID:   input.SessionID,
		UserID:      input.UserID,
		IPHash:      input.IPHash,
		UserAgent:   input.UserAgent,
		Referer:     input.Referer,
		UTMSource:   input.UTMSource,
		UTMMedium:   input.UTMMedium,
		UTMCampaign: input.UTMCampaign,
	}); err != nil {
		return "", err
	}
	return link, nil
}
