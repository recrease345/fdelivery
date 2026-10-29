package main

import (
	"context"
	"errors"
	"fdelivery_orders/internal/api"
	"fdelivery_orders/internal/config"
	"fdelivery_orders/internal/logger"
	"fdelivery_orders/internal/order"
	"fdelivery_orders/internal/storage/pg"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// @title orders
// @BasePath /v1
// @securityDefinitions.apikey X-User-ID in header
func main() {
	// логгер
	slogger, err := logger.New()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	cfg, err := config.LoadCFG()
	if err != nil {
		slogger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pg.NewPool(ctx, *cfg)
	if err != nil {
		slogger.Error("failed to create pool", "error", err)
		os.Exit(1)
	}

	ordersRepo := pg.NewOrderRepo(pool)
	keysRepo := pg.NewKeysRepo(pool)
	outboxRepo := pg.NewOutboxRepo()
	tx := pg.NewTxManager(pool)
	pub := nopPublisher{log: slogger}

	svc := order.NewService(ordersRepo, keysRepo, outboxRepo, tx, &pub, slogger)
	h := api.NewHandler(svc, slogger)

	// маршрутизация
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/orders", h.Create)
	mux.HandleFunc("GET /v1/orders/{id}", h.Get)
	mux.HandleFunc("PATCH /v1/orders/{id}", h.Transit)
	mux.HandleFunc("POST /v1/orders/{id}/cancel", h.Cancel)

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(503)
			return
		}

		w.WriteHeader(200)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`)) // потом тело
	})

	server := http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	slogger.Info("Starting server", "port", server.Addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slogger.Error("server error", "error", err)
		}
	}()

	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slogger.Error("Server stopped forcibly", "error", err)
	}

	pool.Close()
	slogger.Info("Server stopped")
}

type nopPublisher struct{ log *slog.Logger }

func (p *nopPublisher) Publish(_ context.Context, ev order.Event) error {
	p.log.Warn("publisher: nop-stab, event stays in outbox",
		"event", ev)

	return nil
}
