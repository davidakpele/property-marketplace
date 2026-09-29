package search

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/pkg/cache"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

const cacheTTL = 5 * time.Minute

type cachedResult struct {
	Listings []*domain.Listing `json:"listings"`
	Total    int64             `json:"total"`
}

func cacheKey(filters domain.SearchFilters, params pagination.Params) string {
	b, _ := json.Marshal(struct {
		F domain.SearchFilters `json:"f"`
		P pagination.Params    `json:"p"`
	}{F: filters, P: params})
	return fmt.Sprintf("search:%x", md5.Sum(b))
}

func getCached(ctx context.Context, c *cache.Client, key string) ([]*domain.Listing, int64, bool) {
	raw, err := c.Get(ctx, key)
	if err != nil {
		return nil, 0, false
	}
	var result cachedResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, 0, false
	}
	return result.Listings, result.Total, true
}

func setCached(ctx context.Context, c *cache.Client, key string, listings []*domain.Listing, total int64) {
	b, err := json.Marshal(cachedResult{Listings: listings, Total: total})
	if err != nil {
		return
	}
	_ = c.Set(ctx, key, string(b), cacheTTL)
}
