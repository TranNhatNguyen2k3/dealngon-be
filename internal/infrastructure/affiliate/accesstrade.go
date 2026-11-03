package affiliate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dealngon/backend/internal/domain"
)

const defaultAccessTradeAPI = "https://api.accesstrade.vn"
const defaultDeepLinkHost = "https://go.isclix.com"

// AccessTradeProvider generates Accesstrade deep links via API (preferred) or
// go.isclix.com deep_link URL when only the access key is configured.
type AccessTradeProvider struct {
	AccessKey    string
	SiteID       string
	APIBaseURL   string
	DeepLinkHost string
	HTTPClient   *http.Client
}

func NewAccessTrade(accessKey, siteID, apiBaseURL string) *AccessTradeProvider {
	base := strings.TrimRight(strings.TrimSpace(apiBaseURL), "/")
	if base == "" {
		base = defaultAccessTradeAPI
	}
	return &AccessTradeProvider{
		AccessKey:    strings.TrimSpace(accessKey),
		SiteID:       strings.TrimSpace(siteID),
		APIBaseURL:   base,
		DeepLinkHost: defaultDeepLinkHost,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *AccessTradeProvider) Name() string { return "accesstrade" }

func (p *AccessTradeProvider) GenerateLink(ctx context.Context, offer domain.Offer) (string, error) {
	if p.AccessKey == "" {
		return "", domain.NewError("ACCESSTRADE_NOT_CONFIGURED", "Thiếu ACCESSTRADE_ACCESS_KEY", domain.ErrInvalid)
	}
	destination, err := destinationURL(offer)
	if err != nil {
		return "", err
	}
	subID := "offer-" + strconv.FormatUint(offer.ID, 10)
	if p.SiteID != "" {
		link, err := p.createDeepLinkAPI(ctx, destination, subID)
		if err != nil {
			return "", err
		}
		return link, nil
	}
	return p.buildDeepLinkURL(destination, subID), nil
}

func destinationURL(offer domain.Offer) (string, error) {
	raw := strings.TrimSpace(offer.ProductURL)
	if raw == "" {
		raw = strings.TrimSpace(offer.AffiliateURL)
	}
	if raw == "" {
		return "", domain.NewError("AFFILIATE_LINK_MISSING", "Offer chưa có đường dẫn sản phẩm", domain.ErrInvalid)
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", domain.NewError("AFFILIATE_LINK_INVALID", "Đường dẫn sản phẩm không hợp lệ", domain.ErrInvalid)
	}
	return parsed.String(), nil
}

type deepLinkRequest struct {
	URL         string   `json:"url"`
	UTMSource   string   `json:"utm_source,omitempty"`
	UTMMedium   string   `json:"utm_medium,omitempty"`
	UTMCampaign string   `json:"utm_campaign,omitempty"`
	SubIDs      []string `json:"subids,omitempty"`
}

type deepLinkResponse struct {
	Success any    `json:"success"`
	Data    any    `json:"data"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (p *AccessTradeProvider) createDeepLinkAPI(ctx context.Context, destination, subID string) (string, error) {
	endpoint := fmt.Sprintf("%s/v1/publishers/me/sites/%s/deeplinks", p.APIBaseURL, url.PathEscape(p.SiteID))
	body, err := json.Marshal(deepLinkRequest{
		URL:         destination,
		UTMSource:   "dealngon",
		UTMMedium:   "affiliate",
		UTMCampaign: subID,
		SubIDs:      []string{subID},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Token "+p.AccessKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := p.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", domain.NewError("ACCESSTRADE_UNAVAILABLE", "Không gọi được AccessTrade", err)
	}
	defer res.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", domain.NewError("ACCESSTRADE_HTTP_ERROR", fmt.Sprintf("AccessTrade trả lỗi HTTP %d", res.StatusCode), domain.ErrInvalid)
	}
	link, err := extractDeepLink(payload, destination)
	if err != nil {
		return "", err
	}
	return link, nil
}

func extractDeepLink(payload []byte, destination string) (string, error) {
	var raw deepLinkResponse
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", domain.NewError("ACCESSTRADE_BAD_RESPONSE", "Phản hồi AccessTrade không hợp lệ", err)
	}
	candidates := collectLinkCandidates(raw.Data)
	for _, candidate := range candidates {
		if candidate == "" || candidate == destination {
			continue
		}
		if parsed, err := url.Parse(candidate); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			return parsed.String(), nil
		}
	}
	msg := raw.Message
	if msg == "" {
		msg = raw.Error
	}
	if msg == "" {
		msg = "AccessTrade không trả về deep link"
	}
	return "", domain.NewError("ACCESSTRADE_NO_DEEPLINK", msg, domain.ErrInvalid)
}

func collectLinkCandidates(data any) []string {
	var out []string
	switch value := data.(type) {
	case nil:
		return out
	case string:
		out = append(out, value)
	case map[string]any:
		for _, key := range []string{"affiliate_link", "deeplink", "deep_link", "tracking_link", "url", "link"} {
			if text, ok := value[key].(string); ok {
				out = append(out, text)
			}
		}
	case []any:
		for _, item := range value {
			out = append(out, collectLinkCandidates(item)...)
		}
	}
	return out
}

func (p *AccessTradeProvider) buildDeepLinkURL(destination, subID string) string {
	base := strings.TrimRight(p.DeepLinkHost, "/")
	if base == "" {
		base = defaultDeepLinkHost
	}
	u, _ := url.Parse(base + "/deep_link/" + url.PathEscape(p.AccessKey))
	query := u.Query()
	query.Set("url", destination)
	query.Set("utm_source", "dealngon")
	query.Set("utm_medium", "affiliate")
	query.Set("utm_campaign", subID)
	query.Set("aff_sub", subID)
	u.RawQuery = query.Encode()
	return u.String()
}
