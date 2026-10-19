package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// startMetricsServer serves Prometheus metrics on a dedicated port, separate
// from the public API port, so /metrics is never reachable through the ingress.
func startMetricsServer(ctx context.Context, port string, db *sqlx.DB, logger *zap.SugaredLogger) {
	prometheus.MustRegister(collectors.NewDBStatsCollector(db.DB, "main-server"))

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Warnw("metrics server shutdown error", "error", err)
		}
	}()

	logger.Infow("metrics server listening", "port", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Errorw("metrics server stopped", "port", port, "error", err)
	}
}
