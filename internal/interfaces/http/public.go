package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dealngon/backend/internal/application"
	"github.com/dealngon/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) ready(c *gin.Context) {
	if s.PingDB != nil {
		if err := s.PingDB(c.Request.Context()); err != nil {
			fail(c, http.StatusServiceUnavailable, "NOT_READY", "Cơ sở dữ liệu chưa sẵn sàng")
			return
		}
	}
	if s.PingRedis != nil {
		if err := s.PingRedis(c.Request.Context()); err != nil {
			fail(c, http.StatusServiceUnavailable, "NOT_READY", "Redis chưa sẵn sàng")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (s *Server) sitemap(c *gin.Context) {
	body, err := s.Catalog.Sitemap(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", []byte(body))
}

func (s *Server) listProducts(c *gin.Context) {
	items, meta, err := s.Catalog.Search(c.Request.Context(), searchQuery(c))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) search(c *gin.Context) {
	s.listProducts(c)
}

func (s *Server) productByID(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	detail, err := s.Catalog.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, detail, nil)
}

func (s *Server) productBySlug(c *gin.Context) {
	detail, err := s.Catalog.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, detail, nil)
}

func (s *Server) productOffers(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	offers, err := s.Catalog.ListOffers(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, offers, nil)
}

func (s *Server) productHistory(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	history, err := s.Catalog.PriceHistory(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, history, nil)
}

func (s *Server) categories(c *gin.Context) {
	items, err := s.Catalog.ListCategories(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) categoryBySlug(c *gin.Context) {
	page, limit := pageQuery(c)
	category, items, meta, err := s.Catalog.Category(c.Request.Context(), c.Param("slug"), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"category": category, "products": items}, meta)
}

func (s *Server) brands(c *gin.Context) {
	items, err := s.Catalog.ListBrands(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) brandBySlug(c *gin.Context) {
	page, limit := pageQuery(c)
	brand, items, meta, err := s.Catalog.Brand(c.Request.Context(), c.Param("slug"), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"brand": brand, "products": items}, meta)
}

func (s *Server) merchants(c *gin.Context) {
	items, err := s.Catalog.ListMerchants(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) deals(c *gin.Context) {
	query := searchQuery(c)
	if query.Sort == "relevance" {
		query.Sort = "deal"
	}
	items, meta, err := s.Catalog.ListDeals(c.Request.Context(), query)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) dealByID(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	deal, err := s.Catalog.DealByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, deal, nil)
}

func (s *Server) dealBySlug(c *gin.Context) {
	deal, err := s.Catalog.DealBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, deal, nil)
}

func (s *Server) createAlert(c *gin.Context) {
	var body domain.PriceAlert
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	alert, err := s.Catalog.CreateAlert(c.Request.Context(), body)
	if err != nil {
		writeError(c, err)
		return
	}
	created(c, alert)
}

func (s *Server) createEvent(c *gin.Context) {
	var body domain.AnalyticsEvent
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if err := s.Catalog.TrackEvent(c.Request.Context(), body); err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"accepted": true}, nil)
}

func (s *Server) redirect(c *gin.Context) {
	id, okID := pathID(c, "offerId")
	if !okID {
		return
	}
	if s.Limiter != nil {
		key := "rl:redirect:" + c.ClientIP()
		allowed, err := s.Limiter.Allow(c.Request.Context(), key, 30, time.Minute)
		if err == nil && !allowed {
			fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "Quá nhiều yêu cầu, thử lại sau")
			return
		}
	}
	ipHash := ""
	if s.Hasher != nil {
		ipHash = s.Hasher.Hash(c.ClientIP())
	}
	link, err := s.Affiliate.Redirect(c.Request.Context(), application.ClickInput{
		OfferID:     id,
		SessionID:   sessionID(c),
		IPHash:      ipHash,
		UserAgent:   c.Request.UserAgent(),
		Referer:     c.Request.Referer(),
		UTMSource:   c.Query("utm_source"),
		UTMMedium:   c.Query("utm_medium"),
		UTMCampaign: c.Query("utm_campaign"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Redirect(http.StatusFound, link)
}

func searchQuery(c *gin.Context) domain.SearchQuery {
	page, limit := pageQuery(c)
	return domain.SearchQuery{
		Query:        c.Query("q"),
		CategorySlug: c.Query("category"),
		BrandSlug:    c.Query("brand"),
		MerchantSlug: c.Query("merchant"),
		MinPrice:     int64Query(c, "min_price"),
		MaxPrice:     int64Query(c, "max_price"),
		MinDiscount:  floatQuery(c, "min_discount"),
		Sort:         c.DefaultQuery("sort", "relevance"),
		Page:         page,
		Limit:        limit,
	}
}

func pageQuery(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	return domain.NormalizePage(page, limit)
}

func int64Query(c *gin.Context, key string) *int64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &value
}

func floatQuery(c *gin.Context, key string) *float64 {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil
	}
	return &value
}

func pathID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		fail(c, http.StatusBadRequest, "INVALID_ID", "Mã không hợp lệ")
		return 0, false
	}
	return id, true
}

func sessionID(c *gin.Context) string {
	if value := c.Query("session_id"); value != "" {
		return value
	}
	cookie, err := c.Cookie("dn_session")
	if err == nil {
		return cookie
	}
	return ""
}
