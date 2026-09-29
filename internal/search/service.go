package search

import (
	"context"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/cache"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

type Service struct {
	repo  repository.ListingRepository
	cache *cache.Client
}

func NewService(repo repository.ListingRepository, cache *cache.Client) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) Search(ctx context.Context, filters domain.SearchFilters, params pagination.Params) ([]*domain.Listing, int64, error) {
	if filters.RadiusKm != nil && *filters.RadiusKm <= 0 {
		return nil, 0, httpx.NewBadRequest("radius_km must be greater than 0")
	}
	if (filters.Latitude != nil || filters.Longitude != nil) &&
		!(filters.Latitude != nil && filters.Longitude != nil && filters.RadiusKm != nil) {
		return nil, 0, httpx.NewBadRequest("lat, lng and radius_km must all be provided for geo search")
	}

	key := cacheKey(filters, params)

	if s.cache != nil {
		if listings, total, ok := getCached(ctx, s.cache, key); ok {
			return listings, total, nil
		}
	}

	listings, total, err := s.repo.Search(ctx, filters, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, httpx.NewInternal("failed to search listings")
	}

	if s.cache != nil {
		setCached(ctx, s.cache, key, listings, total)
	}

	return listings, total, nil
}
