package subscription

import (
	"context"
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/database"
	db "toy-terrier-telegram/internal/database/sqlc"
	"toy-terrier-telegram/internal/models"
)

// Manager handles subscription operations
type Manager struct {
	store *database.Store
}

// NewManager creates a new subscription manager
func NewManager(sqlDB *sql.DB) *Manager {
	return &Manager{
		store: database.NewStore(sqlDB),
	}
}

// CreateOrUpdateUser creates or updates a user in the database
func (m *Manager) CreateOrUpdateUser(user *tgbotapi.User) error {
	ctx := context.Background()

	// Check if user exists
	_, err := m.store.GetUserByTelegramID(ctx, user.ID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	username := ""
	if user.UserName != "" {
		username = user.UserName
	}

	firstName := ""
	if user.FirstName != "" {
		firstName = user.FirstName
	}

	lastName := ""
	if user.LastName != "" {
		lastName = user.LastName
	}

	languageCode := ""
	if user.LanguageCode != "" {
		languageCode = user.LanguageCode
	}

	if err == sql.ErrNoRows {
		// Create new user
		_, err = m.store.CreateUser(ctx, db.CreateUserParams{
			TelegramID:   user.ID,
			Username:     sql.NullString{String: username, Valid: username != ""},
			FirstName:    sql.NullString{String: firstName, Valid: firstName != ""},
			LastName:     sql.NullString{String: lastName, Valid: lastName != ""},
			LanguageCode: sql.NullString{String: languageCode, Valid: languageCode != ""},
			IsActive:     true,
			IsAdmin:      false,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Info().Int64("telegram_id", user.ID).Msg("Created new user")
	} else {
		// Update existing user
		_, err = m.store.UpdateUser(ctx, db.UpdateUserParams{
			TelegramID:   user.ID,
			Username:     sql.NullString{String: username, Valid: username != ""},
			FirstName:    sql.NullString{String: firstName, Valid: firstName != ""},
			LastName:     sql.NullString{String: lastName, Valid: lastName != ""},
			LanguageCode: sql.NullString{String: languageCode, Valid: languageCode != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		log.Debug().Int64("telegram_id", user.ID).Msg("Updated existing user")
	}

	return nil
}

// UpdateUserActivity updates user's last activity timestamp
func (m *Manager) UpdateUserActivity(telegramID int64) error {
	ctx := context.Background()
	err := m.store.UpdateUserActivity(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to update user activity: %w", err)
	}
	return nil
}

// GetActiveCategories returns all active categories
func (m *Manager) GetActiveCategories() ([]*models.Category, error) {
	ctx := context.Background()

	dbCategories, err := m.store.GetActiveCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	categories := make([]*models.Category, len(dbCategories))
	for i, dbCat := range dbCategories {
		categories[i] = &models.Category{
			ID:         int(dbCat.ID),
			Name:       dbCat.Name,
			Slug:       dbCat.Slug,
			SourceType: dbCat.SourceType,
			SourceURL:  dbCat.SourceUrl,
			IsActive:   dbCat.IsActive,
			CreatedAt:  dbCat.CreatedAt,
			UpdatedAt:  dbCat.UpdatedAt,
		}
	}

	return categories, nil
}

// GetUserSubscriptions returns user's active subscriptions
func (m *Manager) GetUserSubscriptions(telegramID int64) ([]*models.UserSubscription, error) {
	ctx := context.Background()

	// Get user ID first
	user, err := m.store.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	dbSubs, err := m.store.GetUserSubscriptions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	subscriptions := make([]*models.UserSubscription, len(dbSubs))
	for i, dbSub := range dbSubs {
		subscriptions[i] = &models.UserSubscription{
			ID:           int(dbSub.ID),
			UserID:       int(dbSub.UserID),
			CategoryID:   int(dbSub.CategoryID),
			CategoryName: dbSub.CategoryName,
			CategorySlug: dbSub.CategorySlug,
			IsActive:     dbSub.IsActive,
			CreatedAt:    dbSub.CreatedAt,
			UpdatedAt:    dbSub.UpdatedAt,
		}
	}

	return subscriptions, nil
}

// Subscribe subscribes user to a category
func (m *Manager) Subscribe(telegramID int64, categoryID int) error {
	ctx := context.Background()

	// Get user ID first
	user, err := m.store.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	_, err = m.store.CreateSubscription(ctx, db.CreateSubscriptionParams{
		UserID:     user.ID,
		CategoryID: int32(categoryID),
		IsActive:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Info().Int64("telegram_id", telegramID).Int("category_id", categoryID).Msg("User subscribed")
	return nil
}

// Unsubscribe unsubscribes user from a category
func (m *Manager) Unsubscribe(telegramID int64, categoryID int) error {
	ctx := context.Background()

	// Get user ID first
	user, err := m.store.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = m.store.UpdateSubscription(ctx, db.UpdateSubscriptionParams{
		UserID:     user.ID,
		IsActive:   false,
		CategoryID: int32(categoryID),
	})
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	log.Info().Int64("telegram_id", telegramID).Int("category_id", categoryID).Msg("User unsubscribed")
	return nil
}

// IsUserSubscribed checks if user is subscribed to a category
func (m *Manager) IsUserSubscribed(telegramID int64, categoryID int) (bool, error) {
	ctx := context.Background()

	// Get user ID first
	user, err := m.store.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	subscription, err := m.store.GetSubscriptionByUserAndCategory(ctx, db.GetSubscriptionByUserAndCategoryParams{
		UserID:     user.ID,
		CategoryID: int32(categoryID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check subscription: %w", err)
	}

	return subscription.IsActive, nil
}

// UnsubscribeFromAll unsubscribes user from all categories
func (m *Manager) UnsubscribeFromAll(telegramID int64) error {
	ctx := context.Background()

	// Get user ID first
	user, err := m.store.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = m.store.DeactivateUserSubscriptions(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe from all: %w", err)
	}

	log.Info().Int64("telegram_id", telegramID).Msg("User unsubscribed from all categories")
	return nil
}

// GetSystemStats returns system statistics
func (m *Manager) GetSystemStats() (*models.SystemStats, error) {
	ctx := context.Background()

	// Get basic counts
	userCount, err := m.store.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	subscriptionCount, err := m.store.CountSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	// TODO: Add more detailed stats when needed
	stats := &models.SystemStats{
		TotalUsers:         int(userCount),
		ActiveUsers:        int(userCount), // Simplified for now
		TotalSubscriptions: int(subscriptionCount),
		// Add other fields as needed
	}

	return stats, nil
}

// GetRecentUsers returns recently active users
func (m *Manager) GetRecentUsers(limit int) ([]*models.User, error) {
	ctx := context.Background()

	users, err := m.store.GetRecentUsers(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to get recent users: %w", err)
	}

	result := make([]*models.User, len(users))
	for i, user := range users {
		var username, firstName, lastName, languageCode *string
		if user.Username.Valid {
			username = &user.Username.String
		}
		if user.FirstName.Valid {
			firstName = &user.FirstName.String
		}
		if user.LastName.Valid {
			lastName = &user.LastName.String
		}
		if user.LanguageCode.Valid {
			languageCode = &user.LanguageCode.String
		}

		result[i] = &models.User{
			ID:           int(user.ID),
			TelegramID:   user.TelegramID,
			Username:     username,
			FirstName:    firstName,
			LastName:     lastName,
			LanguageCode: languageCode,
			IsActive:     user.IsActive,
			IsAdmin:      user.IsAdmin,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			LastActivity: user.LastActivity,
		}
	}

	return result, nil
}
