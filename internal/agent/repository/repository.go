package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/agent/domain"
)

type AgentRepository interface {
	Create(ctx context.Context, agent *domain.Agent) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error)
	GetByEmail(ctx context.Context, email string) (*domain.Agent, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Agent, int64, error)
	Update(ctx context.Context, agent *domain.Agent) error
	Delete(ctx context.Context, id uuid.UUID) error
}
