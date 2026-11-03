package affiliate

import (
	"log/slog"
	"strings"

	"github.com/dealngon/backend/internal/domain"
)

type Config struct {
	Provider   string
	AccessKey  string
	SiteID     string
	APIBaseURL string
}

// NewProvider selects Accesstrade when configured; otherwise mock for local/dev.
func NewProvider(cfg Config) domain.AffiliateProvider {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	hasKey := strings.TrimSpace(cfg.AccessKey) != ""
	switch provider {
	case "accesstrade":
		if !hasKey {
			slog.Warn("AFFILIATE_PROVIDER=accesstrade nhưng thiếu ACCESSTRADE_ACCESS_KEY; dùng mock")
			return MockProvider{}
		}
		slog.Info("affiliate provider", "name", "accesstrade", "mode", accessTradeMode(cfg.SiteID))
		return NewAccessTrade(cfg.AccessKey, cfg.SiteID, cfg.APIBaseURL)
	case "mock", "":
		if provider == "" && hasKey {
			slog.Info("affiliate provider", "name", "accesstrade", "mode", accessTradeMode(cfg.SiteID))
			return NewAccessTrade(cfg.AccessKey, cfg.SiteID, cfg.APIBaseURL)
		}
		slog.Info("affiliate provider", "name", "mock")
		return MockProvider{}
	default:
		slog.Warn("AFFILIATE_PROVIDER không hỗ trợ; dùng mock", "provider", cfg.Provider)
		return MockProvider{}
	}
}

func accessTradeMode(siteID string) string {
	if strings.TrimSpace(siteID) != "" {
		return "api"
	}
	return "deep_link_url"
}
