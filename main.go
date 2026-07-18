package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yenikule/config"
	"yenikule/handlers"
)

func main() {
	// Load config from environment (no heap allocation per-request)
	cfg := config.Load()

	// Build a single ServeMux — zero global state beyond this
	mux := http.NewServeMux()

	// Static file handler: serves css/, js/, images/ and *.html
	handlers.RegisterStatic(mux, cfg)

	// Telegram webhook handler (only registered when token is set)
	if cfg.TelegramToken != "" {
		handlers.RegisterTelegram(mux, cfg)
		log.Printf("[telegram] webhook registered at /webhook/%s", cfg.WebhookPath)
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,

		// Tight timeouts → no goroutine leak per idle connection
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,

		// Keep the default buffer sizes small
		MaxHeaderBytes: 1 << 13, // 8 KB
	}

	// Graceful shutdown on SIGINT / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("[server] listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] fatal: %v", err)
		}
	}()

	<-quit
	log.Println("[server] shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[server] shutdown error: %v", err)
	}
	log.Println("[server] stopped")
}
