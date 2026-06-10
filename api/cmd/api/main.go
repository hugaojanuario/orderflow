package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"orderflow/api/internal/http/handler"
	"orderflow/api/internal/http/router"
	"orderflow/api/internal/metrics"
	"orderflow/api/internal/queue"
	"orderflow/api/internal/repository"
	"orderflow/api/internal/services"
	"orderflow/api/pkg/config"
	"orderflow/api/pkg/database"
)

// injetados em build time via -ldflags
var (
	version = "dev"
	commit  = "none"
)

func main() {
	seed := flag.Bool("seed", false, "popula o banco com o cardápio e o usuário admin")
	flag.Parse()

	cfg, err := config.LoadDotEnv()
	if err != nil {
		setupLogger("info")
		slog.Error("erro ao carregar a configuração", slog.String("component", "api"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	setupLogger(cfg.LOG_LEVEL)

	db, err := database.Connect(cfg.DATABASE_URL)
	if err != nil {
		slog.Error("erro ao conectar no banco", slog.String("component", "api"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if *seed {
		if err := repository.Seed(db); err != nil {
			slog.Error("erro ao executar o seed", slog.String("component", "api"), slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("seed executado com sucesso", slog.String("component", "api"))
		return
	}

	redisClient, err := database.ConnectRedis(cfg.REDIS_URL)
	if err != nil {
		slog.Error("erro ao conectar no redis", slog.String("component", "api"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	if cfg.APP_ENV == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// repositories
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	cache := repository.NewCache(redisClient)
	publisher := queue.NewPublisher(redisClient)

	// services
	authService := services.NewAuthService(userRepo, cfg.JWT_SECRET)
	orderService := services.NewOrderService(orderRepo, menuRepo, publisher, cache)
	menuService := services.NewMenuService(menuRepo)
	statsService := services.NewStatsService(statsRepo)

	// controllers
	authController := handler.NewAuthController(authService)
	orderController := handler.NewOrderController(orderService)
	menuController := handler.NewMenuController(menuService)
	statsController := handler.NewStatsController(statsService)
	healthController := handler.NewHealthController(db, redisClient)
	versionController := handler.NewVersionController(version, commit)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	poller := metrics.NewStatusPoller(orderRepo, 10*time.Second)
	go poller.Run(ctx)

	engine := router.SetupRouter(authController, orderController, menuController, statsController, healthController, versionController, cfg.JWT_SECRET)

	srv := &http.Server{
		Addr:    ":" + cfg.PORT,
		Handler: engine,
	}

	go func() {
		slog.Info("api iniciada",
			slog.String("component", "api"),
			slog.String("port", cfg.PORT),
			slog.String("version", version),
			slog.String("commit", commit),
			slog.String("env", cfg.APP_ENV))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("erro ao subir o servidor", slog.String("component", "api"), slog.String("error", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()

	slog.Info("sinal recebido, iniciando graceful shutdown", slog.String("component", "api"))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("erro no graceful shutdown", slog.String("component", "api"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("api encerrada", slog.String("component", "api"))
}

func setupLogger(level string) {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
	slog.SetDefault(logger)
}
