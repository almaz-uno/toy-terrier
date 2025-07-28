package models

import "time"

// Forum represents a forum
type Forum struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	BaseURL    string      `json:"base_url"`
	IsActive   bool        `json:"is_active"`
	CreatedAt  time.Time   `json:"created_at"`
	Categories []*Category `json:"categories,omitempty"`
}

// ScrapingMetrics represents scraping metrics
type ScrapingMetrics struct {
	TotalPosts      int            `json:"total_posts"`
	PostsLast24h    int            `json:"posts_last_24h"`
	LastScrapingAt  *time.Time     `json:"last_scraping_at"`
	PostsByCategory map[string]int `json:"posts_by_category"`
}
