package models

import (
	"time"
)

// SubscriptionWithCategory represents subscription with category details
type SubscriptionWithCategory struct {
	Subscription
	Category Category `json:"category"`
}

// UserWithSubscriptions represents user with their subscriptions
type UserWithSubscriptions struct {
	User
	Subscriptions []SubscriptionWithCategory `json:"subscriptions"`
}

// NotificationWithDetails represents notification with user and topic details
type NotificationWithDetails struct {
	Notification
	User  User       `json:"user"`
	Topic ForumTopic `json:"topic"`
}

// NotificationQueue represents notification queue item
type NotificationQueue struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	TopicID     int       `json:"topic_id"`
	MessageText string    `json:"message_text"`
	Priority    int       `json:"priority"`
	ScheduledAt time.Time `json:"scheduled_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// BulkNotification represents bulk notification request
type BulkNotification struct {
	CategoryID  int    `json:"category_id"`
	TopicID     int    `json:"topic_id"`
	MessageText string `json:"message_text"`
	UserIDs     []int  `json:"user_ids,omitempty"`
	ExcludeIDs  []int  `json:"exclude_ids,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent    int `json:"total_sent"`
	TotalFailed  int `json:"total_failed"`
	TotalPending int `json:"total_pending"`
	TotalToday   int `json:"total_today"`
	RateLimited  int `json:"rate_limited"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int `json:"total_users"`
	ActiveUsers   int `json:"active_users"`
	AdminUsers    int `json:"admin_users"`
	NewUsersToday int `json:"new_users_today"`
	ActiveToday   int `json:"active_today"`
}

// TopicStats represents topic statistics
type TopicStats struct {
	TotalTopics    int `json:"total_topics"`
	TopicsToday    int `json:"topics_today"`
	NotifiedTopics int `json:"notified_topics"`
	PendingTopics  int `json:"pending_topics"`
}

// SystemStats represents overall system statistics
type SystemStats struct {
	TotalUsers          int               `json:"total_users"`
	ActiveUsers         int               `json:"active_users"`
	TotalSubscriptions  int               `json:"total_subscriptions"`
	TotalTopics         int               `json:"total_topics"`
	TopicsToday         int               `json:"topics_today"`
	TotalNotifications  int               `json:"total_notifications"`
	NotificationsToday  int               `json:"notifications_today"`
	FailedNotifications int               `json:"failed_notifications"`
	Users               UserStats         `json:"users"`
	Notifications       NotificationStats `json:"notifications"`
	Topics              TopicStats        `json:"topics"`
	LastScrapedAt       *time.Time        `json:"last_scraped_at"`
	UptimeStart         time.Time         `json:"uptime_start"`
}
