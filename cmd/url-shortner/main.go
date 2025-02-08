package main

import (
	"github.com/linemk/url_shortner/internal/http-server/handlers/redirect"
	"log/slog"
	"net/http"
	"os"

	"github.com/linemk/url_shortner/internal/config"
	"github.com/linemk/url_shortner/internal/http-server/handlers/url/save"
	mvLogger "github.com/linemk/url_shortner/internal/http-server/middleware/logger"
	"github.com/linemk/url_shortner/internal/lib/logger/handlers/slogpretty"
	"github.com/linemk/url_shortner/internal/lib/logger/sl"
	"github.com/linemk/url_shortner/internal/storage/sqlite"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// создаем файл конфигурации
	cfg := config.MustLoad()

	// создаем логгер
	log := setupLogger(cfg.Env)

	log.Info("starting server", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	//инициализируем дб
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mvLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortner", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
	})
	router.Get("/url/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.TimeOut,
		WriteTimeout: cfg.HTTPServer.TimeOut,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Info("shutting down server")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	color.NoColor = false

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
