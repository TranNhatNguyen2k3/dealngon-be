package affiliate

import (
	"context"
	"net/url"

	"github.com/dealngon/backend/internal/domain"
)

type MockProvider struct{}

func (MockProvider) GenerateLink(ctx context.Context, offer domain.Offer) (string, error) {
	raw := offer.ProductURL
	if raw == "" {
		raw = offer.AffiliateURL
	}
	if raw == "" {
		return "", domain.NewError("AFFILIATE_LINK_MISSING", "Offer chưa có đường dẫn sản phẩm", domain.ErrInvalid)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domain.NewError("AFFILIATE_LINK_INVALID", "Đường dẫn sản phẩm không hợp lệ", domain.ErrInvalid)
	}
	query := parsed.Query()
	query.Set("utm_source", "dealngon")
	query.Set("utm_medium", "affiliate")
	query.Set("aff_sub", "offer-"+itoa(offer.ID))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func itoa(id uint64) string {
	if id == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = byte('0' + id%10)
		id /= 10
	}
	return string(buf[i:])
}
