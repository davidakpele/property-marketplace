package domain

import (
	"time"

	"github.com/google/uuid"
)

type Listing struct {
	ID          uuid.UUID
	Title       string
	Description string
	Price       float64
	Type        ListingType
	Bedrooms    int
	Address     string
	Latitude    float64
	Longitude   float64
	AgentID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SearchFilters struct {
	Type         *ListingType
	MinPrice     *float64
	MaxPrice     *float64
	Bedrooms     *int
	Latitude     *float64
	Longitude    *float64
	RadiusKm     *float64
}
