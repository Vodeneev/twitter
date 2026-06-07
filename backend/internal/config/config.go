package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env              string
	HTTPAddr         string
	DatabaseURL      string
	DBConnectTimeout time.Duration
	ShutdownTimeout  time.Duration
	CORSOrigins      []string
	CookieSecure     bool
	AdminEmails      []string

	SiteName    string
	SiteBaseURL string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPTLSMode  string

	S3Endpoint         string
	S3ExternalEndpoint string
	S3Region           string
	S3AccessKeyID      string
	S3SecretAccessKey  string
	S3Bucket           string
	S3PublicBaseURL    string
	S3UsePathStyle     bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:              getEnv("APP_ENV", "development"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		DBConnectTimeout: getDuration("DB_CONNECT_TIMEOUT", 10*time.Second),
		ShutdownTimeout:  getDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		CORSOrigins:      splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		CookieSecure:     getBool("COOKIE_SECURE", false),
		AdminEmails:      splitCSV(os.Getenv("ADMIN_EMAILS")),

		SiteName:    getEnv("SITE_NAME", "Yapper"),
		SiteBaseURL: strings.TrimRight(getEnv("SITE_BASE_URL", "http://localhost:3000/ru"), "/"),

		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getInt("SMTP_PORT", 1025),
		SMTPUsername: os.Getenv("SMTP_USER"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     getEnv("SMTP_FROM", "Yapper <no-reply@yapper.local>"),
		SMTPTLSMode:  getEnv("SMTP_TLS_MODE", "none"),

		S3Endpoint:         os.Getenv("S3_ENDPOINT"),
		S3ExternalEndpoint: os.Getenv("S3_EXTERNAL_ENDPOINT"),
		S3Region:           getEnv("S3_REGION", "auto"),
		S3AccessKeyID:      os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey:  os.Getenv("S3_SECRET_ACCESS_KEY"),
		S3Bucket:           os.Getenv("S3_BUCKET"),
		S3PublicBaseURL:    os.Getenv("S3_PUBLIC_BASE_URL"),
		S3UsePathStyle:     getBool("S3_USE_PATH_STYLE", true),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
