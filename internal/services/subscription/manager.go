package subscription

import (
	"database/sql"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/models"
)

// Manager handles subscription operations
type Manager struct {
	db *sql.DB
}

// NewManager creates a new subscription manager
func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// CreateOrUpdateUser creates or updates a user in the database
func (m *Manager) CreateOrUpdateUser(user *tgbotapi.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, language_code, last_activity)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (telegram_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			language_code = EXCLUDED.language_code,
			last_activity = NOW(),
			updated_at = NOW()
	`

	var username, firstName, lastName, languageCode *string
	if user.UserName != "" {
		username = &user.UserName
	}
	if user.FirstName != "" {
		firstName = &user.FirstName
	}
	if user.LastName != "" {
		lastName = &user.LastName
	}
	if user.LanguageCode != "" {
		languageCode = &user.LanguageCode
	}

	_, err := m.db.Exec(query, user.ID, username, firstName, lastName, languageCode)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}

	log.Info().
		Int64("telegram_id", user.ID).
		Str("username", user.UserName).
		Msg("User created/updated")

	return nil
}

// UpdateUserActivity updates user's last activity time
func (m *Manager) UpdateUserActivity(telegramID int64) error {
	query := `UPDATE users SET last_activity = NOW() WHERE telegram_id = $1`
	_, err := m.db.Exec(query, telegramID)
	return err
}

// Subscribe subscribes a user to a category
func (m *Manager) Subscribe(userID, categoryID int) error {
	// First get the actual user ID from telegram_id
	var dbUserID int
	err := m.db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", userID).Scan(&dbUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	query := `
		INSERT INTO subscriptions (user_id, category_id, is_active)
		VALUES ($1, $2, true)
		ON CONFLICT (user_id, category_id)
		DO UPDATE SET is_active = true, updated_at = NOW()
	`

	_, err = m.db.Exec(query, dbUserID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Info().
		Int("user_id", dbUserID).
		Int("category_id", categoryID).
		Msg("User subscribed to category")

	return nil
}

// Unsubscribe unsubscribes a user from a category
func (m *Manager) Unsubscribe(userID, categoryID int) error {
	var dbUserID int
	err := m.db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", userID).Scan(&dbUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	query := `
		UPDATE subscriptions
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND category_id = $2
	`

	_, err = m.db.Exec(query, dbUserID, categoryID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	log.Info().
		Int("user_id", dbUserID).
		Int("category_id", categoryID).
		Msg("User unsubscribed from category")

	return nil
}

// UnsubscribeFromAll unsubscribes a user from all categories
func (m *Manager) UnsubscribeFromAll(userID int) error {
	var dbUserID int
	err := m.db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", userID).Scan(&dbUserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	query := `
		UPDATE subscriptions
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1
	`

	_, err = m.db.Exec(query, dbUserID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from all: %w", err)
	}

	log.Info().
		Int("user_id", dbUserID).
		Msg("User unsubscribed from all categories")

	return nil
}

// IsUserSubscribed checks if user is subscribed to a category
func (m *Manager) IsUserSubscribed(userID, categoryID int) (bool, error) {
	var dbUserID int
	err := m.db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", userID).Scan(&dbUserID)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	var isActive bool
	query := `
		SELECT is_active FROM subscriptions
		WHERE user_id = $1 AND category_id = $2
	`

	err = m.db.QueryRow(query, dbUserID, categoryID).Scan(&isActive)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check subscription: %w", err)
	}

	return isActive, nil
}

// GetActiveCategories returns all active categories
func (m *Manager) GetActiveCategories() ([]interface{}, error) {
	query := `
		SELECT id, name, slug, source_type, source_url
		FROM categories
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []interface{}
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.SourceType, &category.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetUserSubscriptions returns user's active subscriptions
func (m *Manager) GetUserSubscriptions(userID int) ([]models.Category, error) {
	var dbUserID int
	err := m.db.QueryRow("SELECT id FROM users WHERE telegram_id = $1", userID).Scan(&dbUserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	query := `
		SELECT c.id, c.name, c.slug, c.source_type, c.source_url
		FROM categories c
		JOIN subscriptions s ON c.id = s.category_id
		WHERE s.user_id = $1 AND s.is_active = true
		ORDER BY c.name
	`

	rows, err := m.db.Query(query, dbUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.SourceType, &category.SourceURL)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// GetSystemStats returns system statistics
func (m *Manager) GetSystemStats() (*models.SystemStats, error) {
	stats := &models.SystemStats{
		UptimeStart: time.Now(), // This should be set when the app starts
	}

	// Get user stats
	err := m.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM users
	`).Scan(&stats.Users.TotalUsers, &stats.Users.ActiveUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get topic stats
	err = m.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as today
		FROM forum_topics
	`).Scan(&stats.Topics.TotalTopics, &stats.Topics.TopicsToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic stats: %w", err)
	}

	// Get notification stats
	err = m.db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'sent') as sent,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as today
		FROM notifications
	`).Scan(&stats.Notifications.TotalSent, &stats.Notifications.TotalFailed,
		&stats.Notifications.TotalPending, &stats.Notifications.TotalToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	return stats, nil
}

// GetRecentUsers returns recent users
func (m *Manager) GetRecentUsers(limit int) ([]models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name,
			   is_active, last_activity
		FROM users
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := m.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.TelegramID, &user.Username,
			&user.FirstName, &user.LastName, &user.IsActive, &user.LastActivity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
