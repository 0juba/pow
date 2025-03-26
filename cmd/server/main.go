package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := DefaultConfig()
	deps, err := New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	if err := deps.Server.Start(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down gracefully...")

	cancel()

	if err := deps.Server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	if err := deps.SessionStorage.Close(); err != nil {
		log.Printf("Error closing session storage: %v", err)
	}
	if err := deps.ChallengeStorage.Close(); err != nil {
		log.Printf("Error closing challenge storage: %v", err)
	}

	log.Println("Shutdown complete")
}
