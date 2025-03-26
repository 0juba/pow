package pow

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/0juba/pow/pkg/pow"
)

type Option func(*Config)

func WithDifficulty(d uint8) Option {
	return func(c *Config) {
		c.Difficulty = d
	}
}

func WithTTL(t time.Duration) Option {
	return func(c *Config) {
		c.TTL = t
	}
}

type HashcashPOW struct {
	config Config
}

func NewHashcashPOW(opts ...Option) *HashcashPOW {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}
	return &HashcashPOW{
		config: config,
	}
}

func (h *HashcashPOW) GenerateChallenge(resource string) (*ChallengeItem, error) {
	challenge, err := pow.GenerateChallenge(resource, h.config.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("generate challenge: %w", err)
	}

	if challenge == nil {
		return nil, fmt.Errorf("generate challenge: unexpected nil value")
	}

	chItem := &ChallengeItem{
		ID:        uuid.New(),
		ExpiresAt: time.Now().UTC().Add(h.config.TTL),
		Ch:        *challenge,
	}

	return chItem, nil
}

func (h *HashcashPOW) VerifySolution(solution Solution) (bool, error) {
	if time.Since(solution.ChallengeItem.ExpiresAt) > h.config.TTL {
		return false, fmt.Errorf("verify solution: challengeItem expired")
	}

	isValid, err := pow.VerifySolution(pow.Solution{
		Ch:    solution.ChallengeItem.Ch,
		Nonce: solution.Nonce,
	})
	if err != nil {
		return false, fmt.Errorf("verify solution: %w", err)
	}

	return isValid, nil
}
