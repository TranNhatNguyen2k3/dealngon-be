package affiliate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dealngon/backend/internal/domain"
)

func TestAccessTradeAPIDeepLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token test-key" {
			t.Fatalf("auth %s", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/v1/publishers/me/sites/site-1/deeplinks" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": 1,
			"data": map[string]string{
				"affiliate_link": "https://go.isclix.com/deep_link/abc?url=https%3A%2F%2Fshopee.vn%2Fp",
			},
		})
	}))
	defer server.Close()

	provider := NewAccessTrade("test-key", "site-1", server.URL)
	provider.HTTPClient = server.Client()
	link, err := provider.GenerateLink(context.Background(), domain.Offer{
		ID: 9, ProductURL: "https://shopee.vn/p", Status: domain.StatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(link, "go.isclix.com") {
		t.Fatalf("link %s", link)
	}
}

func TestAccessTradeDeepLinkURLFallback(t *testing.T) {
	provider := NewAccessTrade("pub-key", "", "")
	link, err := provider.GenerateLink(context.Background(), domain.Offer{
		ID: 3, ProductURL: "https://www.lazada.vn/products/x.html",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(link, "go.isclix.com/deep_link/pub-key") {
		t.Fatalf("link %s", link)
	}
	if !strings.Contains(link, "aff_sub=offer-3") {
		t.Fatalf("missing sub id in %s", link)
	}
}

func TestNewProviderSelectsAccessTrade(t *testing.T) {
	p := NewProvider(Config{AccessKey: "k", SiteID: "s"})
	if p.Name() != "accesstrade" {
		t.Fatalf("got %s", p.Name())
	}
	p = NewProvider(Config{Provider: "mock", AccessKey: "k"})
	if p.Name() != "mock" {
		t.Fatalf("got %s", p.Name())
	}
}
