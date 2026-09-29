package handler

import (
	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
)

type CreateListingRequest struct {
	Title       string  `json:"title"       validate:"required,min=3,max=500" example:"3-Bed Flat in Lekki"`
	Description string  `json:"description"                                   example:"Spacious and modern flat"`
	Price       float64 `json:"price"       validate:"required,gt=0"           example:"1500000"`
	Type        string  `json:"type"        validate:"required,oneof=rent sale shortlet" example:"rent"`
	Bedrooms    int     `json:"bedrooms"    validate:"gte=0"                   example:"3"`
	Address     string  `json:"address"     validate:"required,min=3,max=500" example:"14 Admiralty Way, Lekki Phase 1"`
	Latitude    float64 `json:"latitude"    validate:"required,gte=-90,lte=90"   example:"6.4281"`
	Longitude   float64 `json:"longitude"   validate:"required,gte=-180,lte=180" example:"3.4219"`
	AgentID     string  `json:"agent_id"    validate:"required,uuid"           example:"a1000000-0000-0000-0000-000000000001"`
}

type UpdateListingRequest struct {
	Title       string  `json:"title"       validate:"required,min=3,max=500" example:"Updated 3-Bed Flat"`
	Description string  `json:"description"                                   example:"Updated description"`
	Price       float64 `json:"price"       validate:"required,gt=0"           example:"1800000"`
	Type        string  `json:"type"        validate:"required,oneof=rent sale shortlet" example:"sale"`
	Bedrooms    int     `json:"bedrooms"    validate:"gte=0"                   example:"3"`
	Address     string  `json:"address"     validate:"required,min=3,max=500" example:"14 Admiralty Way, Lekki Phase 1"`
	Latitude    float64 `json:"latitude"    validate:"required,gte=-90,lte=90"   example:"6.4281"`
	Longitude   float64 `json:"longitude"   validate:"required,gte=-180,lte=180" example:"3.4219"`
	AgentID     string  `json:"agent_id"    validate:"required,uuid"           example:"a1000000-0000-0000-0000-000000000001"`
}

type SearchRequest struct {
	Type     string   `form:"type"      validate:"omitempty,oneof=rent sale shortlet"`
	MinPrice *float64 `form:"min_price" validate:"omitempty,gt=0"`
	MaxPrice *float64 `form:"max_price" validate:"omitempty,gt=0"`
	Bedrooms *int     `form:"bedrooms"  validate:"omitempty,gte=0"`
	Lat      *float64 `form:"lat"       validate:"omitempty,gte=-90,lte=90"`
	Lng      *float64 `form:"lng"       validate:"omitempty,gte=-180,lte=180"`
	RadiusKm *float64 `form:"radius_km" validate:"omitempty,gt=0"`
}

func (r *SearchRequest) ToFilters() domain.SearchFilters {
	f := domain.SearchFilters{
		MinPrice:  r.MinPrice,
		MaxPrice:  r.MaxPrice,
		Bedrooms:  r.Bedrooms,
		Latitude:  r.Lat,
		Longitude: r.Lng,
		RadiusKm:  r.RadiusKm,
	}
	if r.Type != "" {
		t := domain.ListingType(r.Type)
		f.Type = &t
	}
	return f
}

func (r *CreateListingRequest) AgentUUID() uuid.UUID {
	id, _ := uuid.Parse(r.AgentID)
	return id
}

func (r *UpdateListingRequest) AgentUUID() uuid.UUID {
	id, _ := uuid.Parse(r.AgentID)
	return id
}
