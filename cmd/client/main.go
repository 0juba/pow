package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/0juba/pow/pkg/client"
	"github.com/0juba/pow/pkg/pow"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "Server address")
	count := flag.Int("n", 1, "Number of quotes to get")
	flag.Parse()

	if *count < 1 {
		log.Fatal("Number of quotes must be positive")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	cfg := client.DefaultConfig()
	cfg.Addr = *addr
	c := client.New(cfg)

	if err := c.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer c.Close()

	challengeResp, err := c.RequestChallenge()
	if err != nil {
		log.Fatalf("Failed to get challengeResp: %v", err)
	}

	fmt.Printf("Got challengeResp:\n")
	fmt.Printf("  ID: %s\n", challengeResp.ChallengeID)
	fmt.Printf("  Stamp: %d\n", challengeResp.Stamp)

	fmt.Printf("\nSolving POW challengeResp...\n")

	challenge, _, err := pow.Parse(challengeResp.Stamp)
	if err != nil {
		log.Fatal("parse challenge stamp: %w", err)
	}

	if challenge == nil {
		log.Fatal("got unexpected stamp nil value")
	}

	solution := pow.SolveChallenge(*challenge)

	fmt.Printf("Found solution: %d\n", solution.Nonce)

	fmt.Printf("\nSubmitting solution...\n")
	sessionID, err := c.SubmitSolution(challengeResp.ChallengeID, solution.Nonce)
	if err != nil {
		log.Fatalf("Failed to submit solution: %v", err)
	}

	fmt.Printf("Got session ID: %s\n", sessionID)

	fmt.Printf("\nGetting quotes...\n")
	for i := 0; i < *count; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			quote, err := c.GetQuote(sessionID)
			if err != nil {
				log.Fatalf("Failed to get quote: %v", err)
			}
			fmt.Printf("%d. %s\n", i+1, quote)
		}
	}
}
