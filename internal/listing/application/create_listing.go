package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/internal/listing/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
)

type CreateListingInput struct {
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

type CreateListingUseCase struct {
	repo repository.ListingRepository
}

func NewCreateListingUseCase(repo repository.ListingRepository) *CreateListingUseCase {
	return &CreateListingUseCase{repo: repo}
}

func (uc *CreateListingUseCase) Execute(ctx context.Context, in CreateListingInput) (*domain.Listing, error) {
	if !in.Type.IsValid() {
		return nil, httpx.NewBadRequest("invalid listing type; must be rent, sale or shortlet")
	}

	now := time.Now().UTC()
	l := &domain.Listing{
		ID:          uuid.New(),
		Title:       in.Title,
		Description: in.Description,
		Price:       in.Price,
		Type:        in.Type,
		Bedrooms:    in.Bedrooms,
		Address:     in.Address,
		Latitude:    in.Latitude,
		Longitude:   in.Longitude,
		AgentID:     in.AgentID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.repo.Create(ctx, l); err != nil {
		return nil, httpx.NewInternal("failed to create listing")
	}
	return l, nil
}
