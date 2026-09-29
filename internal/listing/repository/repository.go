package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
)

type ListingRepository interface {
	Create(ctx context.Context, listing *domain.Listing) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Listing, int64, error)
	Search(ctx context.Context, filters domain.SearchFilters, limit, offset int) ([]*domain.Listing, int64, error)
	Update(ctx context.Context, listing *domain.Listing) error
	Delete(ctx context.Context, id uuid.UUID) error
}
