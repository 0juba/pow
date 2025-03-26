package memory

import (
	"fmt"
	"math/rand"
	"sync"
)

type QuoteStorage struct {
	mu sync.Mutex
}

func NewQuoteStorage() *QuoteStorage {
	return &QuoteStorage{}
}

func (s *QuoteStorage) GetRandom() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(quotes) == 0 {
		return "", fmt.Errorf("get random quote: no quotes available")
	}

	return quotes[rand.Intn(len(quotes))], nil
}
