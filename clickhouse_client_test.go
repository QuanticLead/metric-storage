package main

import (
	"testing"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru/v2"
)

func TestClickHouseMetricsStore_CacheEviction(t *testing.T) {
	// Create an LRU cache with size 2
	cache, err := lru.New[uuid.UUID, bool](2)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}

	store := &ClickHouseMetricsStore{
		metadataCache: cache,
	}

	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()

	// Add id1 and id2
	store.metadataCache.Add(id1, true)
	store.metadataCache.Add(id2, true)

	// Verify both exist
	if !store.metadataCache.Contains(id1) || !store.metadataCache.Contains(id2) {
		t.Errorf("expected both id1 and id2 to be in cache")
	}

	// Add id3, which should evict id1 (since id1 was added first and not accessed)
	store.metadataCache.Add(id3, true)

	if store.metadataCache.Contains(id1) {
		t.Errorf("expected id1 to be evicted from cache")
	}
	if !store.metadataCache.Contains(id2) || !store.metadataCache.Contains(id3) {
		t.Errorf("expected id2 and id3 to be in cache")
	}
}
