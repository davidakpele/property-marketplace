package agent

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/davidakpele/property-marketplace/internal/agent/domain"
	"github.com/davidakpele/property-marketplace/internal/agent/repository"
	"github.com/davidakpele/property-marketplace/pkg/httpx"
	"github.com/davidakpele/property-marketplace/pkg/pagination"
)

type CreateAgentInput struct {
	Name   string
	Email  string
	Phone  string
	Agency string
}

type UpdateAgentInput struct {
	Name   string
	Email  string
	Phone  string
	Agency string
}

type Service struct {
	repo repository.AgentRepository
}

func NewService(repo repository.AgentRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateAgentInput) (*domain.Agent, error) {
	existing, err := s.repo.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, httpx.NewInternal("failed to check agent email")
	}
	if existing != nil {
		return nil, httpx.NewConflict("agent with this email already exists")
	}

	now := time.Now().UTC()
	a := &domain.Agent{
		ID:        uuid.New(),
		Name:      in.Name,
		Email:     in.Email,
		Phone:     in.Phone,
		Agency:    in.Agency,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, httpx.NewInternal("failed to create agent")
	}
	return a, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return nil, httpx.NewNotFound("agent not found")
		}
		return nil, httpx.NewInternal("failed to get agent")
	}
	return a, nil
}

func (s *Service) List(ctx context.Context, params pagination.Params) ([]*domain.Agent, int64, error) {
	agents, total, err := s.repo.List(ctx, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, httpx.NewInternal("failed to list agents")
	}
	return agents, total, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateAgentInput) (*domain.Agent, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return nil, httpx.NewNotFound("agent not found")
		}
		return nil, httpx.NewInternal("failed to get agent")
	}

	if in.Email != a.Email {
		existing, err := s.repo.GetByEmail(ctx, in.Email)
		if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
			return nil, httpx.NewInternal("failed to check agent email")
		}
		if existing != nil {
			return nil, httpx.NewConflict("email already in use")
		}
	}

	a.Name = in.Name
	a.Email = in.Email
	a.Phone = in.Phone
	a.Agency = in.Agency
	a.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, a); err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return nil, httpx.NewNotFound("agent not found")
		}
		return nil, httpx.NewInternal("failed to update agent")
	}
	return a, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return httpx.NewNotFound("agent not found")
		}
		return httpx.NewInternal("failed to delete agent")
	}
	return nil
}
