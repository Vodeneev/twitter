package main

import (
	"context"
	"errors"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/Vodeneev/twitter/backend/internal/auth"
	"github.com/Vodeneev/twitter/backend/internal/config"
	"github.com/Vodeneev/twitter/backend/internal/db"
	"github.com/Vodeneev/twitter/backend/internal/dm"
	apihttp "github.com/Vodeneev/twitter/backend/internal/http"
	"github.com/Vodeneev/twitter/backend/internal/mail"
	"github.com/Vodeneev/twitter/backend/internal/notifications"
	"github.com/Vodeneev/twitter/backend/internal/realtime"
	"github.com/Vodeneev/twitter/backend/internal/social"
	"github.com/Vodeneev/twitter/backend/internal/storage"
	"github.com/Vodeneev/twitter/backend/internal/yaps"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.New(ctx, cfg.DatabaseURL, cfg.DBConnectTimeout)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("db pool ready")

	mailer := buildMailer(cfg, logger)
	authMailer := auth.VerificationMailer{M: mailer, From: cfg.SMTPFrom, SiteName: cfg.SiteName}

	users := auth.NewUserRepository(pool)
	if len(cfg.AdminEmails) > 0 {
		n, err := users.GrantAdminByEmails(ctx, cfg.AdminEmails)
		if err != nil {
			logger.Error("failed to apply ADMIN_EMAILS", "error", err)
			os.Exit(1)
		}
		logger.Info("applied admin list", "configured", len(cfg.AdminEmails), "updated_users", n)
	}

	authSvc := auth.NewService(
		users,
		auth.NewSessionRepository(pool),
		auth.NewVerificationRepository(pool),
		auth.NewPasswordResetRepository(pool),
		authMailer,
		cfg.SiteBaseURL,
	)

	photoURL := func(string) string { return "" }
	var storageClient *storage.Client
	if cfg.S3Bucket != "" {
		sc, err := storage.New(ctx, storage.Config{
			Endpoint:         cfg.S3Endpoint,
			ExternalEndpoint: cfg.S3ExternalEndpoint,
			Region:           cfg.S3Region,
			AccessKeyID:      cfg.S3AccessKeyID,
			SecretAccessKey:  cfg.S3SecretAccessKey,
			Bucket:           cfg.S3Bucket,
			PublicBaseURL:    cfg.S3PublicBaseURL,
			UsePathStyle:     cfg.S3UsePathStyle,
		})
		if err != nil {
			logger.Error("failed to init s3 storage", "error", err)
			os.Exit(1)
		}
		storageClient = sc
		photoURL = sc.PublicURL
		logger.Info("s3 storage configured", "endpoint", cfg.S3Endpoint, "bucket", cfg.S3Bucket)
	} else {
		logger.Warn("S3_BUCKET not set, media uploads disabled")
	}

	hub := realtime.NewHub()

	deps := apihttp.Deps{
		Pool:          pool,
		AuthService:   authSvc,
		Users:         users,
		Yaps:          yaps.NewRepository(pool, photoURL),
		Social:        social.NewRepository(pool),
		Notifications: notifications.NewRepository(pool, photoURL),
		DM:            dm.NewRepository(pool, photoURL),
		Hub:           hub,
		Storage:       storageClient,
		PhotoURL:      photoURL,
		SiteName:      cfg.SiteName,
		SiteBaseURL:   cfg.SiteBaseURL,
		CORSOrigins:   cfg.CORSOrigins,
		CookieSecure:  cfg.CookieSecure,
	}

	srv := &stdhttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apihttp.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api server starting", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			logger.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutdown signal received, draining connections")
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped cleanly")
}

func buildMailer(cfg *config.Config, logger *slog.Logger) mail.Mailer {
	if cfg.SMTPHost == "" {
		logger.Warn("SMTP_HOST not set, using null mailer (emails will be logged, not delivered)")
		return mail.NullSender{}
	}
	logger.Info("smtp configured", "host", cfg.SMTPHost, "port", cfg.SMTPPort, "tls", cfg.SMTPTLSMode)
	return mail.NewSMTPSender(mail.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
		TLSMode:  cfg.SMTPTLSMode,
	})
}
