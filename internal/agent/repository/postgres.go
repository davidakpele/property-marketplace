package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/davidakpele/property-marketplace/internal/agent/domain"
)

type postgresAgentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAgentRepository(db *pgxpool.Pool) AgentRepository {
	return &postgresAgentRepository{db: db}
}

func (r *postgresAgentRepository) Create(ctx context.Context, a *domain.Agent) error {
	q := `INSERT INTO agents (id, name, email, phone, agency, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, q, a.ID, a.Name, a.Email, a.Phone, a.Agency, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	return nil
}

func (r *postgresAgentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	q := `SELECT id, name, email, phone, agency, created_at, updated_at
	      FROM agents WHERE id = $1`
	row := r.db.QueryRow(ctx, q, id)
	return scanAgent(row)
}

func (r *postgresAgentRepository) GetByEmail(ctx context.Context, email string) (*domain.Agent, error) {
	q := `SELECT id, name, email, phone, agency, created_at, updated_at
	      FROM agents WHERE email = $1`
	row := r.db.QueryRow(ctx, q, email)
	return scanAgent(row)
}

func (r *postgresAgentRepository) List(ctx context.Context, limit, offset int) ([]*domain.Agent, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM agents`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count agents: %w", err)
	}

	q := `SELECT id, name, email, phone, agency, created_at, updated_at
	      FROM agents ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []*domain.Agent
	for rows.Next() {
		a := &domain.Agent{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Phone, &a.Agency, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, a)
	}
	return agents, total, rows.Err()
}

func (r *postgresAgentRepository) Update(ctx context.Context, a *domain.Agent) error {
	q := `UPDATE agents SET name=$1, email=$2, phone=$3, agency=$4, updated_at=$5 WHERE id=$6`
	ct, err := r.db.Exec(ctx, q, a.Name, a.Email, a.Phone, a.Agency, a.UpdatedAt, a.ID)
	if err != nil {
		return fmt.Errorf("update agent: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postgresAgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM agents WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanAgent(row pgx.Row) (*domain.Agent, error) {
	a := &domain.Agent{}
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.Phone, &a.Agency, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("scan agent: %w", err)
	}
	return a, nil
}
