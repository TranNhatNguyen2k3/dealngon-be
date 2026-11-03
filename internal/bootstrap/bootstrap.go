package bootstrap

import (
	"context"
	"log/slog"

	"github.com/dealngon/backend/internal/application"
	"github.com/dealngon/backend/internal/config"
	"github.com/dealngon/backend/internal/domain"
	"github.com/dealngon/backend/internal/infrastructure/affiliate"
	"github.com/dealngon/backend/internal/infrastructure/auth"
	"github.com/dealngon/backend/internal/infrastructure/matcher"
	"github.com/dealngon/backend/internal/infrastructure/merchant"
	"github.com/dealngon/backend/internal/infrastructure/postgres"
	rediscache "github.com/dealngon/backend/internal/infrastructure/redis"
	httpapi "github.com/dealngon/backend/internal/interfaces/http"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config *config.Config
	Router *gin.Engine
	Pool   *pgxpool.Pool
	Redis  *rediscache.Client
	Sync   *application.SyncService
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return nil, err
	}
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	cache, err := rediscache.Connect(ctx, cfg.RedisAddr)
	if err != nil {
		pool.Close()
		return nil, err
	}

	products := postgres.NewProductRepo(pool)
	variants := postgres.NewVariantRepo(pool)
	offers := postgres.NewOfferRepo(pool)
	history := postgres.NewHistoryRepo(pool)
	deals := postgres.NewDealRepo(pool)
	brands := postgres.NewBrandRepo(pool)
	categories := postgres.NewCategoryRepo(pool)
	merchants := postgres.NewMerchantRepo(pool)
	coupons := postgres.NewCouponRepo(pool)
	seo := postgres.NewSEORepo(pool)
	alerts := postgres.NewAlertRepo(pool)
	clicks := postgres.NewClickRepo(pool)
	admins := postgres.NewAdminRepo(pool)
	searches := postgres.NewSearchLogRepo(pool)
	audits := postgres.NewAuditRepo(pool)
	events := postgres.NewEventRepo(pool)

	catalog := &application.CatalogService{
		Products: products, Variants: variants, Offers: offers, History: history, Deals: deals,
		Brands: brands, Categories: categories, Merchants: merchants, Coupons: coupons, SEO: seo,
		Alerts: alerts, Searches: searches, Events: events, Cache: cache, FrontendURL: cfg.FrontendURL,
	}
	seedMerchants := make([]domain.Merchant, 0, len(merchant.Directory()))
	for _, info := range merchant.Directory() {
		seedMerchants = append(seedMerchants, domain.Merchant{
			Name: info.Name, Slug: info.Slug, Website: info.Website,
			AffiliateEnabled: true, Status: domain.StatusActive, QualityScore: info.QualityScore,
		})
	}
	syncService := &application.SyncService{
		Providers: merchant.AllProviders(), Matcher: matcher.New(variants),
		Brands: brands, Categories: categories, Merchants: merchants, Products: products,
		Variants: variants, Offers: offers, History: history, Deals: deals, Cache: cache,
		SeedMerchants: seedMerchants, CategoryLabels: merchant.CategoryNames(),
	}
	affiliateService := &application.AffiliateService{
		Offers: offers, Clicks: clicks, Provider: affiliate.MockProvider{},
	}
	issuer := auth.NewIssuer(cfg.JWTSecret, cfg.IPHashSalt)
	adminService := &application.AdminService{
		Admins: admins, Products: products, Variants: variants, Offers: offers, Brands: brands,
		Categories: categories, Merchants: merchants, Deals: deals, Coupons: coupons, SEO: seo,
		Clicks: clicks, Audits: audits, Tokens: issuer, Sync: syncService, Catalog: catalog,
	}

	if err := adminService.Ensure(ctx, cfg.AdminEmail, cfg.AdminName, cfg.AdminPassword); err != nil {
		pool.Close()
		_ = cache.Close()
		return nil, err
	}
	count, err := products.Count(ctx)
	if err != nil {
		pool.Close()
		_ = cache.Close()
		return nil, err
	}
	if count == 0 {
		slog.Info("seeding catalog from mock providers")
		if err := syncService.Run(ctx); err != nil {
			pool.Close()
			_ = cache.Close()
			return nil, err
		}
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := httpapi.NewRouter(&httpapi.Server{
		Catalog: catalog, Affiliate: affiliateService, Admin: adminService,
		Tokens: issuer, Limiter: cache, Hasher: issuer,
		PingDB: pool.Ping, PingRedis: cache.Ping,
	})
	return &App{Config: &cfg, Router: router, Pool: pool, Redis: cache, Sync: syncService}, nil
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}
