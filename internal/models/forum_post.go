package models

import (
	"time"
)

// ForumPost represents a parsed forum post
type ForumPost struct {
	ID          int       `json:"id"`
	CategoryID  int       `json:"category_id"`
	TopicID     string    `json:"topic_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Content     string    `json:"content"`
	URL         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	ScrapedAt   time.Time `json:"scraped_at"`
	PublishedAt time.Time `json:"published_at"`
	ReplyCount  int       `json:"reply_count"`
	ViewCount   int       `json:"view_count"`
	Images      []string  `json:"images"`
	Hash        string    `json:"hash"`
}

// FacebookPost represents a parsed Facebook post
type FacebookPost struct {
	ID          string    `json:"id"`
	Message     string    `json:"message"`
	Story       string    `json:"story"`
	Link        string    `json:"link"`
	Picture     string    `json:"picture"`
	CreatedTime time.Time `json:"created_time"`
	Type        string    `json:"type"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	Shares      int       `json:"shares"`
	Hash        string    `json:"hash"`
}

// ParsedContent represents any parsed content
type ParsedContent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "forum_post", "facebook_post"
	CategoryID  int                    `json:"category_id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Author      string                 `json:"author"`
	URL         string                 `json:"url"`
	PublishedAt time.Time              `json:"published_at"`
	Images      []string               `json:"images"`
	Metadata    map[string]interface{} `json:"metadata"`
	Hash        string                 `json:"hash"`
}

// ContentType constants
const (
	ContentTypeForumPost    = "forum_post"
	ContentTypeFacebookPost = "facebook_post"
)
