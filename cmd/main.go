package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/newaccg/todo-api/internal/config"
	"github.com/newaccg/todo-api/internal/handler"
	"github.com/newaccg/todo-api/internal/infrastructure/crypto"
	"github.com/newaccg/todo-api/internal/infrastructure/jwt"
	"github.com/newaccg/todo-api/internal/middleware"
	"github.com/newaccg/todo-api/internal/repository"
	"github.com/newaccg/todo-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig("internal/config/config.json")
	if err != nil {
		slog.Error(
			"could not load config",
			"error", err,
		)
		os.Exit(1)
	}

	crypto := crypto.NewCrypto()
	repo := repository.NewRepository(&cfg.DB, crypto)
	if err := repo.LoadDB(); err != nil {
		slog.Error(
			"could not load config",
			"error", err,
		)
		os.Exit(2)
	}
	defer repo.UnloadDB()

	jwt := jwt.NewJWT(cfg.Jwt.ExpirationTime.Duration, cfg.Jwt.Secret, cfg.ValueNames.JwtUserID)
	midware := middleware.NewMiddleware(cfg.Jwt.HeaderName, jwt, cfg.ValueNames.JwtUserID, cfg.RateLimiting.RefillPerSecond, cfg.RateLimiting.RefreshDuration.Duration)
	svc := service.NewService(repo, jwt, &cfg.Jwt)
	h := handler.NewHandler(svc, midware, &cfg.ValueNames.Url, cfg.ValueNames.JwtUserID, &cfg.RateLimiting.BucketSizes)

	mux := h.RegisterRoutes()
	err = http.ListenAndServe(cfg.ServerAddress, mux)
	slog.Error(
		"server shut down",
		"error", err,
	)
	os.Exit(3)
}
