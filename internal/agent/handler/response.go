package handler

import (
	"time"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/agent/domain"
)

type AgentResponse struct {
	ID        uuid.UUID `json:"id"         example:"a1000000-0000-0000-0000-000000000001"`
	Name      string    `json:"name"       example:"Ada Okafor"`
	Email     string    `json:"email"      example:"ada@realty.ng"`
	Phone     string    `json:"phone"      example:"+2348011111111"`
	Agency    string    `json:"agency"     example:"Realty NG"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DataResponse struct {
	Data AgentResponse `json:"data"`
}

type DataListResponse struct {
	Data       []AgentResponse `json:"data"`
	Pagination interface{}     `json:"pagination"`
}

type ErrorResponse struct {
	Message string `json:"message" example:"agent not found"`
}

type ValidationErrorResponse struct {
	Errors []ValidationFieldError `json:"errors"`
}

type ValidationFieldError struct {
	Field   string `json:"field"   example:"email"`
	Message string `json:"message" example:"must be a valid email address"`
}

func toResponse(a *domain.Agent) AgentResponse {
	return AgentResponse{
		ID:        a.ID,
		Name:      a.Name,
		Email:     a.Email,
		Phone:     a.Phone,
		Agency:    a.Agency,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func toResponseList(agents []*domain.Agent) []AgentResponse {
	result := make([]AgentResponse, len(agents))
	for i, a := range agents {
		result[i] = toResponse(a)
	}
	return result
}
