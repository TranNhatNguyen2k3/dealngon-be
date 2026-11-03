package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv                string
	Port                  string
	DatabaseURL           string
	RedisAddr             string
	JWTSecret             string
	FrontendURL           string
	AdminEmail            string
	AdminPassword         string
	AdminName             string
	WorkerInterval        time.Duration
	IPHashSalt            string
	AffiliateProvider     string
	AccessTradeAccessKey  string
	AccessTradeSiteID     string
	AccessTradeAPIBaseURL string
}

func Load() Config {
	host := getenv("DATABASE_HOST", "localhost")
	port := getenv("DATABASE_PORT", "5432")
	user := getenv("DATABASE_USER", "postgres")
	password := getenv("DATABASE_PASSWORD", "postgres")
	name := getenv("DATABASE_NAME", "dealngon")
	redisHost := getenv("REDIS_HOST", "localhost")
	redisPort := getenv("REDIS_PORT", "6379")
	interval := getenv("WORKER_INTERVAL", "15m")
	parsed, err := time.ParseDuration(interval)
	if err != nil || parsed <= 0 {
		parsed = 15 * time.Minute
	}
	return Config{
		AppEnv:                getenv("APP_ENV", "development"),
		Port:                  getenv("SERVER_PORT", "8080"),
		DatabaseURL:           fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name),
		RedisAddr:             redisHost + ":" + redisPort,
		JWTSecret:             getenv("JWT_SECRET", "change_me"),
		FrontendURL:           getenv("FRONTEND_URL", "http://localhost:5173"),
		AdminEmail:            getenv("ADMIN_EMAIL", "admin@dealngon.local"),
		AdminPassword:         getenv("ADMIN_PASSWORD", "admin123456"),
		AdminName:             getenv("ADMIN_NAME", "DealNgon Admin"),
		WorkerInterval:        parsed,
		IPHashSalt:            getenv("IP_HASH_SALT", "dev-salt"),
		AffiliateProvider:     getenv("AFFILIATE_PROVIDER", ""),
		AccessTradeAccessKey:  getenv("ACCESSTRADE_ACCESS_KEY", ""),
		AccessTradeSiteID:     getenv("ACCESSTRADE_SITE_ID", ""),
		AccessTradeAPIBaseURL: getenv("ACCESSTRADE_API_BASE_URL", "https://api.accesstrade.vn"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
