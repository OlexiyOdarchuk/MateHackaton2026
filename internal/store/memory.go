package store

import (
	"sync"
	"time"

	"SuperAdds/internal/ai"
)

// StoredAd keeps the latest summary and generation result for a given page.
type StoredAd struct {
	PageID      string       `json:"page_id"`
	UserContext string       `json:"user_context"`
	BrandInfo   ai.BrandInfo `json:"brand_info"`
	Competitors []string     `json:"competitors"`
	Summary     string       `json:"summary"`
	ImageURL    string       `json:"image_url"`
	CreatedAt   time.Time    `json:"created_at"`
}

// MemoryStore is a very light in-memory repository for generated ads.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]StoredAd
}

// NewMemoryStore initializes the store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]StoredAd),
	}
}

// Save stores or replaces the ad data for a page.
func (s *MemoryStore) Save(ad StoredAd) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ad.CreatedAt = time.Now().UTC()
	s.items[ad.PageID] = ad
}

// Get returns the stored ad for a page, if any.
func (s *MemoryStore) Get(pageID string) (StoredAd, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ad, ok := s.items[pageID]
	return ad, ok
}
