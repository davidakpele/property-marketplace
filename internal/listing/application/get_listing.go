package application

import (
	"context"
	"errors"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetListingUseCase struct {
	repo repository.ListingRepository
}

func NewGetListingUseCase(repo repository.ListingRepository) *GetListingUseCase {
	return &GetListingUseCase{repo: repo}
}

func (uc *GetListingUseCase) Execute(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	l, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.NewNotFound("listing not found")
		}
		return nil, httpx.NewInternal("failed to get listing")
	}
	return l, nil
}
