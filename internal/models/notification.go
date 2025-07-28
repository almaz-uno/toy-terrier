package models

import (
	"time"
)

// NotificationTemplate represents a notification message template
type NotificationTemplate struct {
	Type     string                 `json:"type"`
	Template string                 `json:"template"`
	Data     map[string]interface{} `json:"data"`
}

// TelegramMessage represents a Telegram message to be sent
type TelegramMessage struct {
	ChatID      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ParseMode   string      `json:"parse_mode,omitempty"`
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	NotificationID int    `json:"notification_id"`
	Success        bool   `json:"success"`
	MessageID      *int   `json:"message_id,omitempty"`
	Error          string `json:"error,omitempty"`
	RetryAfter     *int   `json:"retry_after,omitempty"`
}

// BatchNotificationResult represents the result of sending batch notifications
type BatchNotificationResult struct {
	TotalSent   int                  `json:"total_sent"`
	TotalFailed int                  `json:"total_failed"`
	Results     []NotificationResult `json:"results"`
	ProcessedAt time.Time            `json:"processed_at"`
}

// QueueMetrics represents notification queue metrics
type QueueMetrics struct {
	PendingCount     int           `json:"pending_count"`
	ProcessingCount  int           `json:"processing_count"`
	FailedCount      int           `json:"failed_count"`
	RateLimitedCount int           `json:"rate_limited_count"`
	AverageWaitTime  time.Duration `json:"average_wait_time"`
	LastProcessedAt  *time.Time    `json:"last_processed_at"`
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID               int    `json:"user_id"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	QuietHoursStart      *int   `json:"quiet_hours_start,omitempty"` // Hour 0-23
	QuietHoursEnd        *int   `json:"quiet_hours_end,omitempty"`   // Hour 0-23
	Language             string `json:"language"`
	Categories           []int  `json:"categories"`
}

// MessageFormat represents different message formats
type MessageFormat struct {
	Text     string `json:"text"`
	HTML     string `json:"html"`
	Markdown string `json:"markdown"`
}

// NotificationPriority constants
const (
	PriorityLow    = 1
	PriorityNormal = 2
	PriorityHigh   = 3
	PriorityUrgent = 4
)

// Parse mode constants
const (
	ParseModeHTML       = "HTML"
	ParseModeMarkdown   = "Markdown"
	ParseModeMarkdownV2 = "MarkdownV2"
)
