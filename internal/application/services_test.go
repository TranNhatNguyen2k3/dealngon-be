package application

import (
	"context"
	"errors"
	"testing"

	"github.com/dealngon/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type fakeProducts struct {
	cards []domain.ProductCard
	query domain.SearchQuery
}

func (f *fakeProducts) FindByID(context.Context, uint64) (*domain.Product, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeProducts) FindBySlug(context.Context, string) (*domain.Product, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeProducts) Search(_ context.Context, query domain.SearchQuery) ([]domain.ProductCard, int64, error) {
	f.query = query
	return f.cards, int64(len(f.cards)), nil
}
func (f *fakeProducts) List(context.Context, int, int) ([]domain.Product, int64, error) {
	return nil, 0, nil
}
func (f *fakeProducts) Create(context.Context, domain.Product) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProducts) Update(context.Context, domain.Product) (*domain.Product, error) {
	return nil, nil
}
func (f *fakeProducts) Delete(context.Context, uint64) error      { return nil }
func (f *fakeProducts) Count(context.Context) (int64, error)      { return 0, nil }
func (f *fakeProducts) ListIDs(context.Context) ([]uint64, error) { return nil, nil }
func (f *fakeProducts) ListSitemap(context.Context) ([]domain.SitemapEntry, error) {
	return nil, nil
}

type fakeSearches struct{ calls int }

func (f *fakeSearches) Create(context.Context, string, int) error {
	f.calls++
	return nil
}

func TestSearchNormalizesLimitAndLogsQuery(t *testing.T) {
	products := &fakeProducts{cards: []domain.ProductCard{{Name: "iPhone 16 128GB"}}}
	logs := &fakeSearches{}
	service := &CatalogService{Products: products, Searches: logs}
	items, meta, err := service.Search(context.Background(), domain.SearchQuery{Query: "iphone", Page: 0, Limit: 500, Sort: "nope"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || meta.Limit != 100 || meta.Page != 1 {
		t.Fatalf("items %d meta %+v", len(items), meta)
	}
	if products.query.Sort != "relevance" {
		t.Fatalf("sort = %s", products.query.Sort)
	}
	if logs.calls != 1 {
		t.Fatalf("search logs = %d", logs.calls)
	}
}

type fakeOffers struct{ offer *domain.Offer }

func (f *fakeOffers) FindByID(context.Context, uint64) (*domain.Offer, error) {
	if f.offer == nil {
		return nil, domain.ErrNotFound
	}
	return f.offer, nil
}
func (f *fakeOffers) ListByProduct(context.Context, uint64) ([]domain.Offer, error) { return nil, nil }
func (f *fakeOffers) List(context.Context, int, int) ([]domain.Offer, int64, error) {
	return nil, 0, nil
}
func (f *fakeOffers) Upsert(context.Context, domain.Offer) (*domain.Offer, error) { return nil, nil }
func (f *fakeOffers) Update(context.Context, domain.Offer) (*domain.Offer, error) { return nil, nil }
func (f *fakeOffers) Delete(context.Context, uint64) error                        { return nil }
func (f *fakeOffers) Count(context.Context) (int64, error)                        { return 0, nil }
func (f *fakeOffers) SetAffiliateURL(context.Context, uint64, string) error       { return nil }

type fakeClicks struct{ saved domain.AffiliateClick }

func (f *fakeClicks) Create(_ context.Context, click domain.AffiliateClick) error {
	f.saved = click
	return nil
}
func (f *fakeClicks) SaveLink(context.Context, uint64, string, string) error { return nil }
func (f *fakeClicks) CountToday(context.Context) (int64, error)              { return 0, nil }

type fakeAffiliate struct{}

func (fakeAffiliate) GenerateLink(context.Context, domain.Offer) (string, error) {
	return "https://merchant.example/item?aff_sub=offer-7", nil
}

func TestAffiliateRedirectRecordsClick(t *testing.T) {
	clicks := &fakeClicks{}
	service := &AffiliateService{
		Offers:   &fakeOffers{offer: &domain.Offer{ID: 7, ProductID: 3, MerchantID: 2, Status: domain.StatusActive, ProductURL: "https://merchant.example/item"}},
		Clicks:   clicks,
		Provider: fakeAffiliate{},
	}
	link, err := service.Redirect(context.Background(), ClickInput{OfferID: 7, IPHash: "abc", SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if link == "" || clicks.saved.OfferID != 7 || clicks.saved.ProductID != 3 || clicks.saved.IPHash != "abc" {
		t.Fatalf("link %s click %+v", link, clicks.saved)
	}
}

func TestAffiliateRedirectMissingOffer(t *testing.T) {
	service := &AffiliateService{Offers: &fakeOffers{}, Clicks: &fakeClicks{}, Provider: fakeAffiliate{}}
	_, err := service.Redirect(context.Background(), ClickInput{OfferID: 9})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

type fakeAdmins struct{ admin *domain.Admin }

func (f *fakeAdmins) Ensure(context.Context, string, string, string) (*domain.Admin, error) {
	return f.admin, nil
}
func (f *fakeAdmins) FindByEmail(context.Context, string) (*domain.Admin, error) {
	if f.admin == nil {
		return nil, domain.ErrNotFound
	}
	return f.admin, nil
}
func (f *fakeAdmins) FindByID(context.Context, uint64) (*domain.Admin, error) { return f.admin, nil }

type fakeTokens struct{}

func (fakeTokens) Issue(uint64, string) (string, error) { return "token", nil }
func (fakeTokens) Parse(string) (uint64, error)         { return 1, nil }

func TestAdminLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	service := &AdminService{
		Admins: &fakeAdmins{admin: &domain.Admin{ID: 1, Email: "admin@dealngon.local", PasswordHash: string(hash)}},
		Tokens: fakeTokens{},
	}
	result, err := service.Login(context.Background(), "admin@dealngon.local", "secret")
	if err != nil || result.Token != "token" {
		t.Fatalf("result %+v err %v", result, err)
	}
	if _, err := service.Login(context.Background(), "admin@dealngon.local", "wrong"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("wrong password err = %v", err)
	}
}
