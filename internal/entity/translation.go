// Package entity defines main entities for business logic (services), data base mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Translation -.
type Translation struct {
	ID          uuid.UUID `json:"id"           example:"550e8400-e29b-41d4-a716-446655440000"`
	Source      string    `json:"source"       example:"auto"`
	Destination string    `json:"destination"  example:"en"`
	Original    string    `json:"original"     example:"текст для перевода"`
	Translation string    `json:"translation"  example:"text for translation"`
	CreatedAt   time.Time `json:"created_at"   example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at"   example:"2024-01-01T00:00:00Z"`
}
