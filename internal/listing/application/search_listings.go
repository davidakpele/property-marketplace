package application

import (
	"context"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

type SearchListingsUseCase struct {
	repo repository.ListingRepository
}

func NewSearchListingsUseCase(repo repository.ListingRepository) *SearchListingsUseCase {
	return &SearchListingsUseCase{repo: repo}
}

func (uc *SearchListingsUseCase) Execute(ctx context.Context, filters domain.SearchFilters, params pagination.Params) ([]*domain.Listing, int64, error) {
	if filters.RadiusKm != nil && *filters.RadiusKm <= 0 {
		return nil, 0, httpx.NewBadRequest("radius_km must be greater than 0")
	}
	if (filters.Latitude != nil || filters.Longitude != nil) &&
		!(filters.Latitude != nil && filters.Longitude != nil && filters.RadiusKm != nil) {
		return nil, 0, httpx.NewBadRequest("lat, lng and radius_km must all be provided together for geo search")
	}

	listings, total, err := uc.repo.Search(ctx, filters, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, httpx.NewInternal("failed to search listings")
	}
	return listings, total, nil
}
