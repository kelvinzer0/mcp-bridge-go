package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelvinzer0/mcp-bridge-go/internal/config"
	"github.com/kelvinzer0/mcp-bridge-go/internal/room"
	"github.com/kelvinzer0/mcp-bridge-go/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("Configuration error: %v", err)
	}

	hub := room.NewHub()
	handler := server.NewHandler(hub)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", handler.HandleNew)
	mux.HandleFunc("/mcp/new", handler.HandleNew)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/ws/extension", handler.HandleWSExtension)
	mux.HandleFunc("/mcp", handler.HandleMCP)

	corsHandler := server.CorsMiddleware(mux)

	httpServer := &http.Server{
		Addr:         cfg.Address(),
		Handler:      corsHandler,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting mcp-bridge server on http://%s", cfg.Address())
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down mcp-bridge server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped cleanly")
}
