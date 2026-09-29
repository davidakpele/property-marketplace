package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
)

type DeleteListingUseCase struct {
	repo repository.ListingRepository
}

func NewDeleteListingUseCase(repo repository.ListingRepository) *DeleteListingUseCase {
	return &DeleteListingUseCase{repo: repo}
}

func (uc *DeleteListingUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	err := uc.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.NewNotFound("listing not found")
		}
		return httpx.NewInternal("failed to delete listing")
	}
	return nil
}
