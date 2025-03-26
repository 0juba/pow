package main

import (
	"context"
	"time"

	"github.com/0juba/pow/internal/pow"
	"github.com/0juba/pow/internal/server"
	"github.com/0juba/pow/internal/storage/memory"
)

type Dependencies struct {
	Server           *server.Server
	SessionStorage   *memory.Storage
	ChallengeStorage *memory.Storage
	QuoteStorage     *memory.QuoteStorage
	POW              pow.POWService
}

type Config struct {
	Addr       string
	Difficulty uint8
}

func DefaultConfig() Config {
	return Config{
		Addr:       ":8080",
		Difficulty: 4,
	}
}

func New(ctx context.Context, cfg Config) (*Dependencies, error) {
	sessionStorage := memory.NewInMemoryStorage(ctx)
	challengeStorage := memory.NewInMemoryStorage(ctx)
	quoteStorage := memory.NewQuoteStorage()
	powSrv := pow.NewHashcashPOW(
		pow.WithDifficulty(cfg.Difficulty),
		pow.WithTTL(time.Minute*5),
	)

	srv := server.NewServer(
		ctx,
		sessionStorage,
		challengeStorage,
		quoteStorage,
		powSrv,
	)

	return &Dependencies{
		Server:           srv,
		SessionStorage:   sessionStorage,
		ChallengeStorage: challengeStorage,
		QuoteStorage:     quoteStorage,
	}, nil
}
