package models

import "time"

// Category represents a forum category
type Category struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Slug       string    `json:"slug"`
	SourceType string    `json:"source_type"` // 'forum', 'facebook'
	SourceURL  string    `json:"source_url"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
