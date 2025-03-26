package pow

import (
	"time"

	"github.com/google/uuid"

	"github.com/0juba/pow/pkg/pow"
)

const (
	defaultDifficulty = 4
	defaultTTL        = 5 * time.Minute
)

type Config struct {
	Difficulty uint8
	TTL        time.Duration
}

func DefaultConfig() Config {
	return Config{
		Difficulty: defaultDifficulty,
		TTL:        defaultTTL,
	}
}

type ChallengeItem struct {
	ID        uuid.UUID
	ExpiresAt time.Time
	Ch        pow.Challenge
}

type Solution struct {
	ChallengeItem ChallengeItem
	Nonce         uint64
}

type POWService interface {
	GenerateChallenge(resource string) (*ChallengeItem, error)
	VerifySolution(solution Solution) (bool, error)
}
