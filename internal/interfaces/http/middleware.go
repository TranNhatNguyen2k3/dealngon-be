package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

const adminIDKey = "adminID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			var buf [12]byte
			_, _ = rand.Read(buf[:])
			id = hex.EncodeToString(buf[:])
		}
		c.Set("requestID", id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("http",
			"request_id", c.GetString("requestID"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}
}

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-XSS-Protection", "0")
		c.Next()
	}
}

func CORS(frontendURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", frontendURL)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		c.Header("Vary", "Origin")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func MaxBody(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

func RateLimit(limiter domain.RateLimiter, limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}
		key := "rl:" + c.ClientIP() + ":" + time.Now().UTC().Format("200601021504")
		allowed, err := limiter.Allow(c.Request.Context(), key, limit, time.Minute)
		if err != nil {
			slog.Error("rate limit", "error", err.Error())
			c.Next()
			return
		}
		if !allowed {
			fail(c, http.StatusTooManyRequests, "RATE_LIMITED", "Quá nhiều yêu cầu, thử lại sau")
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdmin(issuer domain.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Thiếu token")
			c.Abort()
			return
		}
		id, err := issuer.Parse(token)
		if err != nil {
			fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Token không hợp lệ")
			c.Abort()
			return
		}
		c.Set(adminIDKey, id)
		c.Next()
	}
}

func adminID(c *gin.Context) uint64 {
	value, _ := c.Get(adminIDKey)
	id, _ := value.(uint64)
	return id
}
