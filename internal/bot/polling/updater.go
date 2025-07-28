package polling

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/bot/handlers"
	"toy-terrier-telegram/internal/config"
)

// Service handles Telegram bot polling
type Service struct {
	config   *config.Config
	api      *tgbotapi.BotAPI
	handlers *handlers.Handlers

	// Statistics
	requestCount int64
	updateCount  int64
	errorCount   int64
	lastPollTime time.Time
	isRunning    bool
	mu           sync.RWMutex
}

// New creates a new polling service
func New(cfg *config.Config, api *tgbotapi.BotAPI, handlers *handlers.Handlers) *Service {
	return &Service{
		config:   cfg,
		api:      api,
		handlers: handlers,
	}
}

// Start starts the polling service
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	s.isRunning = true
	s.mu.Unlock()

	log.Info().
		Int("timeout", s.config.Telegram.Polling.Timeout).
		Int("limit", s.config.Telegram.Polling.Limit).
		Strs("allowed_updates", s.config.Telegram.Polling.AllowedUpdates).
		Msg("Starting polling service")

	// Configure update request
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = s.config.Telegram.Polling.Timeout
	updateConfig.Limit = s.config.Telegram.Polling.Limit
	updateConfig.AllowedUpdates = s.config.Telegram.Polling.AllowedUpdates

	updates := s.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Polling service stopped by context")
			s.mu.Lock()
			s.isRunning = false
			s.mu.Unlock()
			return ctx.Err()

		case update := <-updates:
			atomic.AddInt64(&s.requestCount, 1)
			s.mu.Lock()
			s.lastPollTime = time.Now()
			s.mu.Unlock()

			if update.UpdateID != 0 {
				atomic.AddInt64(&s.updateCount, 1)
				go s.processUpdate(update)
			}
		}
	}
}

// Stop stops the polling service
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRunning = false
	s.api.StopReceivingUpdates()

	log.Info().Msg("Polling service stopped")
	return nil
}

// processUpdate processes a single update
func (s *Service) processUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&s.errorCount, 1)
			log.Error().
				Interface("panic", r).
				Int("update_id", update.UpdateID).
				Msg("Panic while processing update")
		}
	}()

	if err := s.handlers.HandleUpdate(update); err != nil {
		atomic.AddInt64(&s.errorCount, 1)
		log.Error().
			Err(err).
			Int("update_id", update.UpdateID).
			Msg("Error processing update")
	}
}

// GetStats returns polling statistics
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"is_running":      s.isRunning,
		"request_count":   atomic.LoadInt64(&s.requestCount),
		"update_count":    atomic.LoadInt64(&s.updateCount),
		"error_count":     atomic.LoadInt64(&s.errorCount),
		"last_poll_time":  s.lastPollTime,
		"timeout":         s.config.Telegram.Polling.Timeout,
		"limit":           s.config.Telegram.Polling.Limit,
		"allowed_updates": s.config.Telegram.Polling.AllowedUpdates,
	}
}
