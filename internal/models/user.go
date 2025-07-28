package models

import (
	"time"
)

// User represents a Telegram user
type User struct {
	ID           int       `json:"id" db:"id"`
	TelegramID   int64     `json:"telegram_id" db:"telegram_id"`
	Username     *string   `json:"username" db:"username"`
	FirstName    *string   `json:"first_name" db:"first_name"`
	LastName     *string   `json:"last_name" db:"last_name"`
	LanguageCode *string   `json:"language_code" db:"language_code"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	IsAdmin      bool      `json:"is_admin" db:"is_admin"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
}

// Subscription represents user subscription to a category
type Subscription struct {
	ID         int       `json:"id" db:"id"`
	UserID     int       `json:"user_id" db:"user_id"`
	CategoryID int       `json:"category_id" db:"category_id"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// ForumTopic represents a forum topic or post
type ForumTopic struct {
	ID            int       `json:"id" db:"id"`
	CategoryID    int       `json:"category_id" db:"category_id"`
	Title         string    `json:"title" db:"title"`
	Author        string    `json:"author" db:"author"`
	SourceTopicID string    `json:"source_topic_id" db:"source_topic_id"`
	SourceURL     string    `json:"source_url" db:"source_url"`
	Content       string    `json:"content" db:"content"`
	ReplyCount    int       `json:"reply_count" db:"reply_count"`
	ViewCount     int       `json:"view_count" db:"view_count"`
	LastPostTime  time.Time `json:"last_post_time" db:"last_post_time"`
	ContentHash   string    `json:"content_hash" db:"content_hash"`
	HasImages     bool      `json:"has_images" db:"has_images"`
	ImageURLs     []string  `json:"image_urls" db:"image_urls"`
	IsNotified    bool      `json:"is_notified" db:"is_notified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// Notification represents a notification to be sent to a user
type Notification struct {
	ID                int        `json:"id" db:"id"`
	UserID            int        `json:"user_id" db:"user_id"`
	TopicID           int        `json:"topic_id" db:"topic_id"`
	MessageText       string     `json:"message_text" db:"message_text"`
	Status            string     `json:"status" db:"status"` // 'pending', 'sent', 'failed', 'rate_limited'
	TelegramMessageID *int       `json:"telegram_message_id" db:"telegram_message_id"`
	ErrorMessage      *string    `json:"error_message" db:"error_message"`
	RetryCount        int        `json:"retry_count" db:"retry_count"`
	NextRetryAt       *time.Time `json:"next_retry_at" db:"next_retry_at"`
	SentAt            *time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// UserSetting represents user settings
type UserSetting struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	SettingKey   string    `json:"setting_key" db:"setting_key"`
	SettingValue string    `json:"setting_value" db:"setting_value"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// NotificationStatus constants
const (
	NotificationStatusPending     = "pending"
	NotificationStatusSent        = "sent"
	NotificationStatusFailed      = "failed"
	NotificationStatusRateLimited = "rate_limited"
)

// Source type constants
const (
	SourceTypeForum    = "forum"
	SourceTypeFacebook = "facebook"
)

// User setting keys
const (
	SettingNotificationsEnabled = "notifications_enabled"
	SettingLanguage             = "language"
	SettingQuietHoursStart      = "quiet_hours_start"
	SettingQuietHoursEnd        = "quiet_hours_end"
)
