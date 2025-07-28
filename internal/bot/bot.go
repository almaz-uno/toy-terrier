package bot

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/bot/handlers"
	"toy-terrier-telegram/internal/bot/polling"
	"toy-terrier-telegram/internal/config"
	"toy-terrier-telegram/internal/services/notification"
	"toy-terrier-telegram/internal/services/subscription"
)

// Bot represents the Telegram bot
type Bot struct {
	api          *tgbotapi.BotAPI
	config       *config.Config
	db           *sql.DB
	handlers     *handlers.Handlers
	polling      *polling.Service
	subscription *subscription.Manager
	notification *notification.Service
	startTime    time.Time
}

// New creates a new Telegram bot instance
func New(cfg *config.Config, db *sql.DB) (*Bot, error) {
	// Create Telegram Bot API instance
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	api.Debug = cfg.Telegram.Debug

	log.Info().
		Str("username", api.Self.UserName).
		Int64("id", api.Self.ID).
		Msg("Telegram bot authorized")

	// Initialize services
	subscriptionManager := subscription.NewManager(db)
	notificationService := notification.NewService(cfg, db, api)

	// Initialize handlers
	handlerManager := handlers.New(cfg, db, api, subscriptionManager, notificationService)

	// Initialize polling service
	pollingService := polling.New(cfg, api, handlerManager)

	bot := &Bot{
		api:          api,
		config:       cfg,
		db:           db,
		handlers:     handlerManager,
		polling:      pollingService,
		subscription: subscriptionManager,
		notification: notificationService,
		startTime:    time.Now(),
	}

	return bot, nil
}

// Start starts the bot polling
func (b *Bot) Start(ctx context.Context) error {
	log.Info().Msg("Starting Telegram bot polling")

	// Start notification service
	go func() {
		if err := b.notification.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Notification service error")
		}
	}()

	// Start polling
	return b.polling.Start(ctx)
}

// Stop stops the bot
func (b *Bot) Stop() error {
	log.Info().Msg("Stopping Telegram bot")
	return b.polling.Stop()
}

// GetAPI returns the Telegram Bot API instance
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}

// GetStats returns bot statistics
func (b *Bot) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"uptime":       time.Since(b.startTime).String(),
		"start_time":   b.startTime,
		"bot_username": b.api.Self.UserName,
		"bot_id":       b.api.Self.ID,
		"polling":      b.polling.GetStats(),
	}
}

// SendMessage sends a message to a chat
func (b *Bot) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	return b.api.Send(msg)
}

// SendMessageWithKeyboard sends a message with inline keyboard
func (b *Bot) SendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard
	return b.api.Send(msg)
}

// EditMessage edits an existing message
func (b *Bot) EditMessage(chatID int64, messageID int, text string) (tgbotapi.Message, error) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = tgbotapi.ModeHTML
	return b.api.Send(edit)
}

// AnswerCallbackQuery answers a callback query
func (b *Bot) AnswerCallbackQuery(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Request(callback)
	return err
}

// IsAdmin checks if user is admin
func (b *Bot) IsAdmin(userID int64) bool {
	for _, adminID := range b.config.Telegram.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// BroadcastMessage sends a message to all active subscribers
func (b *Bot) BroadcastMessage(message string) error {
	return b.notification.BroadcastToAllUsers(message)
}

// NotifyCategory sends notifications to users subscribed to a category
func (b *Bot) NotifyCategory(categoryID int, topicID int, message string) error {
	return b.notification.NotifySubscribers(categoryID, topicID, message)
}
