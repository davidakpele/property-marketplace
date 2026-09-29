package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/davidakpele/property-marketplace/internal/listing/domain"
)

type postgresListingRepository struct {
	db *pgxpool.Pool
}

func NewPostgresListingRepository(db *pgxpool.Pool) ListingRepository {
	return &postgresListingRepository{db: db}
}

const listingColumns = `id, title, description, price, type, bedrooms, address, latitude, longitude, agent_id, created_at, updated_at`

func (r *postgresListingRepository) Create(ctx context.Context, l *domain.Listing) error {
	q := `INSERT INTO listings (id, title, description, price, type, bedrooms, address, latitude, longitude, agent_id, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(ctx, q,
		l.ID, l.Title, l.Description, l.Price, l.Type, l.Bedrooms,
		l.Address, l.Latitude, l.Longitude, l.AgentID, l.CreatedAt, l.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create listing: %w", err)
	}
	return nil
}

func (r *postgresListingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Listing, error) {
	q := fmt.Sprintf(`SELECT %s FROM listings WHERE id = $1`, listingColumns)
	row := r.db.QueryRow(ctx, q, id)
	return scanListing(row)
}

func (r *postgresListingRepository) List(ctx context.Context, limit, offset int) ([]*domain.Listing, int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM listings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}

	q := fmt.Sprintf(`SELECT %s FROM listings ORDER BY created_at DESC LIMIT $1 OFFSET $2`, listingColumns)
	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list listings: %w", err)
	}
	defer rows.Close()

	listings, err := collectListings(rows)
	if err != nil {
		return nil, 0, err
	}
	return listings, total, nil
}

func (r *postgresListingRepository) Search(ctx context.Context, f domain.SearchFilters, limit, offset int) ([]*domain.Listing, int64, error) {
	args := []interface{}{}
	conditions := []string{}
	idx := 1

	if f.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", idx))
		args = append(args, string(*f.Type))
		idx++
	}
	if f.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", idx))
		args = append(args, *f.MinPrice)
		idx++
	}
	if f.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", idx))
		args = append(args, *f.MaxPrice)
		idx++
	}
	if f.Bedrooms != nil {
		conditions = append(conditions, fmt.Sprintf("bedrooms = $%d", idx))
		args = append(args, *f.Bedrooms)
		idx++
	}
	if f.Latitude != nil && f.Longitude != nil && f.RadiusKm != nil {
		conditions = append(conditions, fmt.Sprintf(
			"ST_DWithin(location, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			idx, idx+1, idx+2,
		))
		args = append(args, *f.Longitude, *f.Latitude, *f.RadiusKm*1000)
		idx += 3
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM listings %s`, where)
	var total int64
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search results: %w", err)
	}

	args = append(args, limit, offset)
	dataQ := fmt.Sprintf(
		`SELECT %s FROM listings %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		listingColumns, where, idx, idx+1,
	)
	rows, err := r.db.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search listings: %w", err)
	}
	defer rows.Close()

	listings, err := collectListings(rows)
	if err != nil {
		return nil, 0, err
	}
	return listings, total, nil
}

func (r *postgresListingRepository) Update(ctx context.Context, l *domain.Listing) error {
	q := `UPDATE listings
	      SET title=$1, description=$2, price=$3, type=$4, bedrooms=$5, address=$6,
	          latitude=$7, longitude=$8, agent_id=$9, updated_at=$10
	      WHERE id=$11`
	ct, err := r.db.Exec(ctx, q,
		l.Title, l.Description, l.Price, l.Type, l.Bedrooms, l.Address,
		l.Latitude, l.Longitude, l.AgentID, l.UpdatedAt, l.ID,
	)
	if err != nil {
		return fmt.Errorf("update listing: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *postgresListingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.db.Exec(ctx, `DELETE FROM listings WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanListing(row pgx.Row) (*domain.Listing, error) {
	l := &domain.Listing{}
	err := row.Scan(
		&l.ID, &l.Title, &l.Description, &l.Price, &l.Type, &l.Bedrooms,
		&l.Address, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("scan listing: %w", err)
	}
	return l, nil
}

func collectListings(rows pgx.Rows) ([]*domain.Listing, error) {
	var listings []*domain.Listing
	for rows.Next() {
		l := &domain.Listing{}
		if err := rows.Scan(
			&l.ID, &l.Title, &l.Description, &l.Price, &l.Type, &l.Bedrooms,
			&l.Address, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan listing row: %w", err)
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}
