package domain

import (
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	Agency    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
