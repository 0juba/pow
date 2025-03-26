package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GenerateChallenge(resource string, difficulty uint8) (*Challenge, error) {
	saltBytes := make([]byte, 128)
	if _, err := rand.Read(saltBytes); err != nil {
		return &Challenge{}, fmt.Errorf("rand read: %w", err)
	}

	salt := base64.StdEncoding.EncodeToString(saltBytes)

	now := time.Now().UTC()

	return &Challenge{
		Timestamp:  now.Unix(),
		Difficulty: difficulty,
		Resource:   base64.StdEncoding.EncodeToString([]byte(resource)),
		Salt:       salt,
	}, nil
}

func VerifySolution(solution Solution) (bool, error) {
	challengeStr := solution.Ch.Stamp(solution.Nonce)

	hash := sha256.Sum256([]byte(challengeStr))
	hashStr := hex.EncodeToString(hash[:])

	prefix := strings.Repeat("0", int(solution.Ch.Difficulty))

	return hashStr[:solution.Ch.Difficulty] == prefix, nil
}

func SolveChallenge(ch Challenge) Solution {
	targetPrefix := strings.Repeat("0", int(ch.Difficulty))

	for nonce := uint64(0); ; nonce++ {
		stamp := ch.Stamp(nonce)
		hash := sha256.Sum256([]byte(stamp))

		hashStr := hex.EncodeToString(hash[:])

		if hashStr[:ch.Difficulty] == targetPrefix {
			return Solution{
				Ch:    ch,
				Nonce: nonce,
			}
		}
	}
}
