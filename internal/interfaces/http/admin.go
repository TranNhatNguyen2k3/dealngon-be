package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/dealngon/backend/internal/application"
	"github.com/dealngon/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) login(c *gin.Context) {
	if s.Cache != nil {
		allowed, err := s.Cache.Allow(c.Request.Context(), "rl:login:"+c.ClientIP(), 10, time.Minute)
		if err == nil && !allowed {
			fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "Quá nhiều lần đăng nhập")
			return
		}
	}
	var body loginRequest
	if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" || body.Password == "" {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "Cần email và mật khẩu")
		return
	}
	result, err := s.Admin.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, result, nil)
}

func (s *Server) dashboard(c *gin.Context) {
	stats, err := s.Admin.Dashboard(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, stats, nil)
}

func (s *Server) adminProducts(c *gin.Context) {
	page, limit := pageQuery(c)
	items, meta, err := s.Admin.ListProducts(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) saveProduct(c *gin.Context) {
	var body application.ProductWrite
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.Product.ID = id
	}
	saved, err := s.Admin.SaveProduct(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteProduct(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	if err := s.Admin.DeleteProduct(c.Request.Context(), adminID(c), id); err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"deleted": true}, nil)
}

func (s *Server) adminBrands(c *gin.Context) {
	items, err := s.Admin.Brands.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) saveBrand(c *gin.Context) {
	var body domain.Brand
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.ID = id
	}
	saved, err := s.Admin.SaveBrand(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteBrand(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteBrand)
}

func (s *Server) adminCategories(c *gin.Context) {
	items, err := s.Admin.Categories.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) saveCategory(c *gin.Context) {
	var body domain.Category
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.ID = id
	}
	saved, err := s.Admin.SaveCategory(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteCategory(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteCategory)
}

func (s *Server) adminMerchants(c *gin.Context) {
	items, err := s.Admin.Merchants.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, nil)
}

func (s *Server) saveMerchant(c *gin.Context) {
	var body domain.Merchant
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.ID = id
	}
	saved, err := s.Admin.SaveMerchant(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteMerchant(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteMerchant)
}

func (s *Server) adminOffers(c *gin.Context) {
	page, limit := pageQuery(c)
	items, meta, err := s.Admin.ListOffers(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) saveOffer(c *gin.Context) {
	var body domain.Offer
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.ID = id
	}
	saved, err := s.Admin.SaveOffer(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteOffer(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteOffer)
}

func (s *Server) adminDeals(c *gin.Context) {
	page, limit := pageQuery(c)
	items, meta, err := s.Admin.ListDeals(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) saveDeal(c *gin.Context) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	deal, err := s.Admin.UpdateDealStatus(c.Request.Context(), adminID(c), id, body.Status)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, deal, nil)
}

func (s *Server) deleteDeal(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteDeal)
}

func (s *Server) recompute(c *gin.Context) {
	if err := s.Admin.Recompute(c.Request.Context(), adminID(c)); err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"synced": true}, nil)
}

func (s *Server) adminCoupons(c *gin.Context) {
	page, limit := pageQuery(c)
	items, meta, err := s.Admin.ListCoupons(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) saveCoupon(c *gin.Context) {
	var body domain.Coupon
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	if id, okID := optionalID(c); okID {
		body.ID = id
	}
	saved, err := s.Admin.SaveCoupon(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteCoupon(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteCoupon)
}

func (s *Server) adminSEO(c *gin.Context) {
	page, limit := pageQuery(c)
	items, meta, err := s.Admin.ListSEO(c.Request.Context(), page, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, items, meta)
}

func (s *Server) saveSEO(c *gin.Context) {
	var body domain.SEOMetadata
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, "INVALID_BODY", "JSON không hợp lệ")
		return
	}
	saved, err := s.Admin.SaveSEO(c.Request.Context(), adminID(c), body)
	if err != nil {
		writeError(c, err)
		return
	}
	ok(c, saved, nil)
}

func (s *Server) deleteSEO(c *gin.Context) {
	s.deleteByID(c, s.Admin.DeleteSEO)
}

func (s *Server) deleteByID(c *gin.Context, fn func(ctx context.Context, adminID, id uint64) error) {
	id, okID := pathID(c, "id")
	if !okID {
		return
	}
	if err := fn(c.Request.Context(), adminID(c), id); err != nil {
		writeError(c, err)
		return
	}
	ok(c, gin.H{"deleted": true}, nil)
}

func optionalID(c *gin.Context) (uint64, bool) {
	raw := c.Param("id")
	if raw == "" {
		return 0, false
	}
	id, okID := pathID(c, "id")
	return id, okID
}
