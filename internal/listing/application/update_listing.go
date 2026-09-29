package application

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
)

type UpdateListingInput struct {
	Title       string
	Description string
	Price       float64
	Type        domain.ListingType
	Bedrooms    int
	Address     string
	Latitude    float64
	Longitude   float64
	AgentID     uuid.UUID
}

type UpdateListingUseCase struct {
	repo repository.ListingRepository
}

func NewUpdateListingUseCase(repo repository.ListingRepository) *UpdateListingUseCase {
	return &UpdateListingUseCase{repo: repo}
}

func (uc *UpdateListingUseCase) Execute(ctx context.Context, id uuid.UUID, in UpdateListingInput) (*domain.Listing, error) {
	l, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.NewNotFound("listing not found")
		}
		return nil, httpx.NewInternal("failed to get listing")
	}

	if !in.Type.IsValid() {
		return nil, httpx.NewBadRequest("invalid listing type; must be rent, sale or shortlet")
	}

	l.Title = in.Title
	l.Description = in.Description
	l.Price = in.Price
	l.Type = in.Type
	l.Bedrooms = in.Bedrooms
	l.Address = in.Address
	l.Latitude = in.Latitude
	l.Longitude = in.Longitude
	l.AgentID = in.AgentID
	l.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, l); err != nil {
		return nil, httpx.NewInternal("failed to update listing")
	}
	return l, nil
}
