package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID        uuid.UUID
	ExpiresAt *time.Time
	Value     any
}

type Storage struct {
	mu     sync.RWMutex
	items  map[uuid.UUID]*Item
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewInMemoryStorage(ctx context.Context) *Storage {
	ctx, cancel := context.WithCancel(ctx)
	s := &Storage{
		items:  make(map[uuid.UUID]*Item),
		ctx:    ctx,
		cancel: cancel,
	}
	s.wg.Add(1)
	go s.cleanupLoop()
	return s
}

func (s *Storage) Close() error {
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *Storage) cleanupLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(defaultCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.cleanup()
		}
	}
}

func (s *Storage) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for id, item := range s.items {
		if item.ExpiresAt != nil && !item.ExpiresAt.IsZero() && now.After(*item.ExpiresAt) {
			delete(s.items, id)
		}
	}
}

func (s *Storage) Store(ctx context.Context, item *Item) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("store challenge: %w", ctx.Err())
	default:
		s.mu.Lock()
		defer s.mu.Unlock()
		s.items[item.ID] = item
		return nil
	}
}

func (s *Storage) Get(ctx context.Context, id uuid.UUID) (*Item, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("get challenge: %w", ctx.Err())
	default:
		s.mu.RLock()
		defer s.mu.RUnlock()

		item, exists := s.items[id]
		if !exists {
			return nil, fmt.Errorf("get challenge: challenge not found")
		}

		if item.ExpiresAt != nil && time.Now().UTC().After(*item.ExpiresAt) {
			return nil, fmt.Errorf("get challenge: challenge expired")
		}

		return item, nil
	}
}

func (s *Storage) Delete(ctx context.Context, id uuid.UUID) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("delete challenge: %w", ctx.Err())
	default:
		s.mu.Lock()
		defer s.mu.Unlock()

		if _, exists := s.items[id]; !exists {
			return fmt.Errorf("delete challenge: challenge not found")
		}

		delete(s.items, id)
		return nil
	}
}
