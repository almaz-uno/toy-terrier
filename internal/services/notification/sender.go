package notification

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"github.com/almaz-uno/toy-terrier/internal/config"
	"github.com/almaz-uno/toy-terrier/internal/models"
)

// Service handles notification operations
type Service struct {
	config *config.Config
	db     *sql.DB
	api    *tgbotapi.BotAPI
}

// NewService creates a new notification service
func NewService(cfg *config.Config, db *sql.DB, api *tgbotapi.BotAPI) *Service {
	return &Service{
		config: cfg,
		db:     db,
		api:    api,
	}
}

// Start starts the notification service
func (s *Service) Start(ctx context.Context) error {
	log.Info().Msg("Starting notification service")

	ticker := time.NewTicker(s.config.Notification.QueueCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Notification service stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.processPendingNotifications(); err != nil {
				log.Error().Err(err).Msg("Failed to process pending notifications")
			}
		}
	}
}

// processPendingNotifications processes pending notifications
func (s *Service) processPendingNotifications() error {
	query := `
		SELECT n.id, n.user_id, n.message_text, n.retry_count, u.telegram_id
		FROM notifications n
		JOIN users u ON n.user_id = u.id
		WHERE n.status = 'pending'
		   OR (n.status = 'rate_limited' AND n.next_retry_at <= NOW())
		ORDER BY n.created_at
		LIMIT $1
	`

	rows, err := s.db.Query(query, s.config.Notification.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var notificationID, userID int
		var messageText string
		var retryCount int
		var telegramID int64

		err := rows.Scan(&notificationID, &userID, &messageText, &retryCount, &telegramID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan notification")
			continue
		}

		if err := s.sendNotification(notificationID, telegramID, messageText, retryCount); err != nil {
			log.Error().
				Err(err).
				Int("notification_id", notificationID).
				Int64("telegram_id", telegramID).
				Msg("Failed to send notification")
		}

		// Rate limiting
		time.Sleep(time.Second / 30) // Max 30 messages per second
	}

	return nil
}

// sendNotification sends a single notification
func (s *Service) sendNotification(notificationID int, telegramID int64, messageText string, retryCount int) error {
	msg := tgbotapi.NewMessage(telegramID, messageText)
	msg.ParseMode = tgbotapi.ModeHTML

	sentMsg, err := s.api.Send(msg)
	if err != nil {
		// Handle rate limiting
		if apiErr, ok := err.(tgbotapi.Error); ok {
			if apiErr.Code == 429 { // Too Many Requests
				retryAfter := 60 * time.Second // Default retry after 1 minute
				nextRetry := time.Now().Add(retryAfter)

				return s.updateNotificationStatus(notificationID, models.NotificationStatusRateLimited,
					&nextRetry, nil, fmt.Sprintf("Rate limited: %s", err.Error()))
			}
		}

		// Handle other errors
		if retryCount >= s.config.Notification.MaxRetries {
			return s.updateNotificationStatus(notificationID, models.NotificationStatusFailed,
				nil, nil, err.Error())
		}

		// Schedule retry
		nextRetry := time.Now().Add(s.config.Notification.RetryDelay * time.Duration(retryCount+1))
		return s.updateNotificationStatus(notificationID, models.NotificationStatusPending,
			&nextRetry, nil, err.Error())
	}

	// Success
	return s.updateNotificationStatus(notificationID, models.NotificationStatusSent,
		nil, &sentMsg.MessageID, "")
}

// updateNotificationStatus updates notification status
func (s *Service) updateNotificationStatus(notificationID int, status string,
	nextRetryAt *time.Time, messageID *int, errorMsg string,
) error {
	query := `
		UPDATE notifications
		SET status = $1, next_retry_at = $2, telegram_message_id = $3,
		    error_message = $4, retry_count = retry_count + 1, sent_at = CASE WHEN $1 = 'sent' THEN NOW() ELSE sent_at END
		WHERE id = $5
	`

	var errorMsgPtr *string
	if errorMsg != "" {
		errorMsgPtr = &errorMsg
	}

	_, err := s.db.Exec(query, status, nextRetryAt, messageID, errorMsgPtr, notificationID)
	return err
}

// NotifySubscribers sends notifications to users subscribed to a category
func (s *Service) NotifySubscribers(categoryID, topicID int, messageText string) error {
	// Get subscribed users
	query := `
		SELECT DISTINCT u.id, u.telegram_id
		FROM users u
		JOIN subscriptions s ON u.id = s.user_id
		WHERE s.category_id = $1 AND s.is_active = true AND u.is_active = true
	`

	rows, err := s.db.Query(query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to get subscribed users: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		var telegramID int64
		if err := rows.Scan(&userID, &telegramID); err != nil {
			log.Error().Err(err).Msg("Failed to scan user")
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		log.Info().Int("category_id", categoryID).Msg("No subscribers found")
		return nil
	}

	// Create notifications
	return s.createBulkNotifications(userIDs, topicID, messageText)
}

// BroadcastToAllUsers sends a message to all active users
func (s *Service) BroadcastToAllUsers(messageText string) error {
	query := `SELECT id FROM users WHERE is_active = true`

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Error().Err(err).Msg("Failed to scan user ID")
			continue
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		log.Info().Msg("No active users found for broadcast")
		return nil
	}

	return s.createBulkNotifications(userIDs, 0, messageText) // topicID = 0 for broadcast
}

// createBulkNotifications creates multiple notifications at once
func (s *Service) createBulkNotifications(userIDs []int, topicID int, messageText string) error {
	query := `
		INSERT INTO notifications (user_id, topic_id, message_text, status, created_at)
		VALUES ($1, $2, $3, 'pending', NOW())
	`

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("Failed to rollback transaction")
		}
	}()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(userID, topicID, messageText)
		if err != nil {
			log.Error().
				Err(err).
				Int("user_id", userID).
				Msg("Failed to create notification")
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Info().
		Int("count", len(userIDs)).
		Int("topic_id", topicID).
		Msg("Created bulk notifications")

	return nil
}

// GetQueueMetrics returns queue metrics
func (s *Service) GetQueueMetrics() (*models.QueueMetrics, error) {
	metrics := &models.QueueMetrics{}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'rate_limited') as rate_limited,
			COUNT(*) FILTER (WHERE status = 'failed') as failed
		FROM notifications
	`

	err := s.db.QueryRow(query).Scan(&metrics.PendingCount, &metrics.RateLimitedCount, &metrics.FailedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue metrics: %w", err)
	}

	// Get last processed time
	err = s.db.QueryRow(`
		SELECT sent_at FROM notifications
		WHERE status = 'sent'
		ORDER BY sent_at DESC
		LIMIT 1
	`).Scan(&metrics.LastProcessedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last processed time: %w", err)
	}

	return metrics, nil
}
