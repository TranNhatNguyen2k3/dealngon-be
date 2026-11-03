package httpapi

import (
	"context"
	_ "embed"
	"net/http"

	"github.com/dealngon/backend/internal/application"
	"github.com/dealngon/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

//go:embed assets/openapi.yaml
var openAPISpec []byte

//go:embed assets/swagger.html
var swaggerHTML []byte

type Server struct {
	Catalog   *application.CatalogService
	Affiliate *application.AffiliateService
	Admin     *application.AdminService
	Tokens    domain.TokenIssuer
	Limiter   domain.RateLimiter
	Hasher    domain.IPHasher
	PingDB    func(context.Context) error
	PingRedis func(context.Context) error
}

func NewRouter(server *Server) *gin.Engine {
	router := gin.New()
	router.Use(RequestID(), AccessLog(), gin.Recovery(), SecureHeaders(), CORS(server.frontendOrigin()), MaxBody(1<<20), RateLimit(server.Limiter, 120))

	router.GET("/health", server.health)
	router.GET("/ready", server.ready)
	router.GET("/sitemap.xml", server.sitemap)
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
	})
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerHTML)
	})

	api := router.Group("/api/v1")
	api.GET("/products", server.listProducts)
	api.GET("/products/slug/:slug", server.productBySlug)
	api.GET("/products/:id", server.productByID)
	api.GET("/products/:id/offers", server.productOffers)
	api.GET("/products/:id/price-history", server.productHistory)
	api.GET("/search", server.search)
	api.GET("/categories", server.categories)
	api.GET("/categories/:slug", server.categoryBySlug)
	api.GET("/brands", server.brands)
	api.GET("/brands/:slug", server.brandBySlug)
	api.GET("/merchants", server.merchants)
	api.GET("/deals", server.deals)
	api.GET("/deals/slug/:slug", server.dealBySlug)
	api.GET("/deals/:id", server.dealByID)
	api.POST("/price-alerts", server.createAlert)
	api.POST("/events", server.createEvent)
	api.GET("/affiliate/redirect/:offerId", server.redirect)

	api.POST("/admin/login", server.login)
	admin := api.Group("/admin", RequireAdmin(server.Tokens))
	admin.GET("/dashboard", server.dashboard)
	admin.POST("/sync", server.recompute)
	admin.GET("/products", server.adminProducts)
	admin.POST("/products", server.saveProduct)
	admin.PUT("/products/:id", server.saveProduct)
	admin.DELETE("/products/:id", server.deleteProduct)
	admin.GET("/brands", server.adminBrands)
	admin.POST("/brands", server.saveBrand)
	admin.PUT("/brands/:id", server.saveBrand)
	admin.DELETE("/brands/:id", server.deleteBrand)
	admin.GET("/categories", server.adminCategories)
	admin.POST("/categories", server.saveCategory)
	admin.PUT("/categories/:id", server.saveCategory)
	admin.DELETE("/categories/:id", server.deleteCategory)
	admin.GET("/merchants", server.adminMerchants)
	admin.POST("/merchants", server.saveMerchant)
	admin.PUT("/merchants/:id", server.saveMerchant)
	admin.DELETE("/merchants/:id", server.deleteMerchant)
	admin.GET("/offers", server.adminOffers)
	admin.POST("/offers", server.saveOffer)
	admin.PUT("/offers/:id", server.saveOffer)
	admin.DELETE("/offers/:id", server.deleteOffer)
	admin.GET("/deals", server.adminDeals)
	admin.PUT("/deals/:id", server.saveDeal)
	admin.DELETE("/deals/:id", server.deleteDeal)
	admin.GET("/coupons", server.adminCoupons)
	admin.POST("/coupons", server.saveCoupon)
	admin.PUT("/coupons/:id", server.saveCoupon)
	admin.DELETE("/coupons/:id", server.deleteCoupon)
	admin.GET("/seo", server.adminSEO)
	admin.POST("/seo", server.saveSEO)
	admin.DELETE("/seo/:id", server.deleteSEO)
	return router
}

func (s *Server) frontendOrigin() string {
	if s.Catalog != nil && s.Catalog.FrontendURL != "" {
		return s.Catalog.FrontendURL
	}
	return "http://localhost:5173"
}
