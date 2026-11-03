package httpapi

import (
	"errors"
	"net/http"

	"github.com/dealngon/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type envelope struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data"`
	Message   any    `json:"message"`
	Meta      any    `json:"meta,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

func ok(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, envelope{Success: true, Data: data, Message: nil, Meta: meta})
}

func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Success: true, Data: data, Message: nil})
}

func fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, envelope{Success: false, Data: nil, Message: message, ErrorCode: code})
}

func writeError(c *gin.Context, err error) {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		fail(c, statusFor(appErr.Err), appErr.Code, appErr.Message)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		fail(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy dữ liệu")
	case errors.Is(err, domain.ErrUnauthorized):
		fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Chưa xác thực")
	case errors.Is(err, domain.ErrInvalid):
		fail(c, http.StatusBadRequest, "INVALID", "Dữ liệu không hợp lệ")
	case errors.Is(err, domain.ErrConflict):
		fail(c, http.StatusConflict, "CONFLICT", "Dữ liệu bị trùng hoặc đang được sử dụng")
	default:
		fail(c, http.StatusInternalServerError, "INTERNAL", "Có lỗi xảy ra")
	}
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}
