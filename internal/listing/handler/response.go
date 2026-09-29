package handler

import (
	"time"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

type ListingResponse struct {
	ID          uuid.UUID `json:"id"          example:"b1000000-0000-0000-0000-000000000001"`
	Title       string    `json:"title"       example:"3-Bed Flat in Lekki"`
	Description string    `json:"description" example:"Spacious and modern flat"`
	Price       float64   `json:"price"       example:"1500000"`
	Type        string    `json:"type"        example:"rent"`
	Bedrooms    int       `json:"bedrooms"    example:"3"`
	Address     string    `json:"address"     example:"14 Admiralty Way, Lekki Phase 1"`
	Latitude    float64   `json:"latitude"    example:"6.4281"`
	Longitude   float64   `json:"longitude"   example:"3.4219"`
	AgentID     uuid.UUID `json:"agent_id"    example:"a1000000-0000-0000-0000-000000000001"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DataResponse struct {
	Data ListingResponse `json:"data"`
}

type DataListResponse struct {
	Data       []ListingResponse `json:"data"`
	Pagination pagination.Meta   `json:"pagination"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"listing not found"`
}

type ValidationErrorResponse struct {
	Errors []ValidationFieldError `json:"errors"`
}

type ValidationFieldError struct {
	Field   string `json:"field"   example:"price"`
	Message string `json:"message" example:"must be greater than 0"`
}

func toResponse(l *domain.Listing) ListingResponse {
	return ListingResponse{
		ID:          l.ID,
		Title:       l.Title,
		Description: l.Description,
		Price:       l.Price,
		Type:        string(l.Type),
		Bedrooms:    l.Bedrooms,
		Address:     l.Address,
		Latitude:    l.Latitude,
		Longitude:   l.Longitude,
		AgentID:     l.AgentID,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}
}

func toResponseList(listings []*domain.Listing) []ListingResponse {
	result := make([]ListingResponse, len(listings))
	for i, l := range listings {
		result[i] = toResponse(l)
	}
	return result
}
