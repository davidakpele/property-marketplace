package application

import (
	"context"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

type ListListingsUseCase struct {
	repo repository.ListingRepository
}

func NewListListingsUseCase(repo repository.ListingRepository) *ListListingsUseCase {
	return &ListListingsUseCase{repo: repo}
}

func (uc *ListListingsUseCase) Execute(ctx context.Context, params pagination.Params) ([]*domain.Listing, int64, error) {
	listings, total, err := uc.repo.List(ctx, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, httpx.NewInternal("failed to list listings")
	}
	return listings, total, nil
}
