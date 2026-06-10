package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	internalhttp "orderflow/worker/internal/http"
	"orderflow/worker/internal/queue"
	"orderflow/worker/internal/repository"
	"orderflow/worker/internal/services"
	"orderflow/worker/pkg/config"
	"orderflow/worker/pkg/database"
)

// injetados em build time via -ldflags
var (
	version = "dev"
	commit  = "none"
)

func main() {
	cfg, err := config.LoadDotEnv()
	if err != nil {
		setupLogger("info")
		slog.Error("erro ao carregar a configuração", slog.String("component", "worker"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	setupLogger(cfg.LOG_LEVEL)

	db, err := database.Connect(cfg.DATABASE_URL)
	if err != nil {
		slog.Error("erro ao conectar no banco", slog.String("component", "worker"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := database.ConnectRedis(cfg.REDIS_URL)
	if err != nil {
		slog.Error("erro ao conectar no redis", slog.String("component", "worker"), slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	orderRepo := repository.NewOrderRepository(db)
	consumer := queue.NewConsumer(redisClient)
	processor := services.NewProcessor(
		orderRepo,
		consumer.Events(),
		cfg.WORKER_CONCURRENCY,
		time.Duration(cfg.ACCEPT_TIME_SECONDS)*time.Second,
		time.Duration(cfg.PREP_TIME_SECONDS)*time.Second,
		time.Duration(cfg.DELIVERY_TIME_SECONDS)*time.Second,
	)

	slog.Info("worker iniciando",
		slog.String("component", "worker"),
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("env", cfg.APP_ENV))

	// recoloca na fila pedidos que ficaram pelo caminho num reinício anterior;
	// o processamento é idempotente, então eventos duplicados são inofensivos
	recoverPending(ctx, orderRepo, consumer)

	go consumer.Run(ctx)

	sizePoller := queue.NewSizePoller(redisClient, 5*time.Second)
	go sizePoller.Run(ctx)

	srv := internalhttp.NewServer(cfg.PORT, db, redisClient)
	go func() {
		slog.Info("servidor de health/metrics iniciado", slog.String("component", "worker"), slog.String("port", cfg.PORT))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("erro ao subir o servidor de health/metrics", slog.String("component", "worker"), slog.String("error", err.Error()))
			cancel()
		}
	}()

	done := make(chan struct{})
	go func() {
		processor.Run(ctx)
		close(done)
	}()

	<-ctx.Done()

	slog.Info("sinal recebido, iniciando graceful shutdown", slog.String("component", "worker"))

	// espera o processor terminar os jobs em andamento
	<-done

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("erro no graceful shutdown", slog.String("component", "worker"), slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("worker encerrado", slog.String("component", "worker"))
}

func recoverPending(ctx context.Context, repo *repository.OrderRepository, consumer *queue.Consumer) {
	ids, err := repo.ListUnfinished()
	if err != nil {
		slog.Error("erro ao recuperar pedidos pendentes", slog.String("component", "worker"), slog.String("error", err.Error()))
		return
	}
	if len(ids) == 0 {
		return
	}

	slog.Info("recolocando pedidos pendentes na fila", slog.String("component", "worker"), slog.Int("count", len(ids)))

	for _, id := range ids {
		if err := consumer.Publish(ctx, id); err != nil {
			slog.Error("erro ao recolocar pedido na fila",
				slog.String("component", "worker"),
				slog.Int("order_id", id),
				slog.String("error", err.Error()))
		}
	}
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
