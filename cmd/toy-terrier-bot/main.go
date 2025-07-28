package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/bot"
	"toy-terrier-telegram/internal/config"
	"toy-terrier-telegram/internal/database"
	"toy-terrier-telegram/internal/server"
	"toy-terrier-telegram/internal/services/notification"
	"toy-terrier-telegram/internal/services/scraper"
	"toy-terrier-telegram/internal/services/subscription"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Run migrations
	if err := database.Migrate(cfg.Database); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	log.Info().Msg("Application started successfully")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Telegram bot
	telegramBot, err := bot.New(cfg, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bot")
	} // Start bot polling
	go func() {
		if err := telegramBot.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Bot polling error")
		}
	}()

	// Initialize services
	_ = subscription.NewManager(db) // Create API instance for services
	api, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create bot API")
	}

	notificationService := notification.NewService(cfg, db, api)
	scraperService := scraper.NewService(cfg, db, notificationService)

	// Start notification service
	go func() {
		if err := notificationService.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Notification service error")
		}
	}()

	// Start scraper service
	go func() {
		if err := scraperService.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Scraper service error")
		}
	}()

	// Initialize and start HTTP server for admin panel and health checks
	httpServer := server.NewServer(cfg, db, notificationService, scraperService)
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Error().Err(err).Msg("HTTP server error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down application...")

	// Cancel context to stop services
	cancel()

	// Give services time to clean up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	if err := httpServer.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Error shutting down HTTP server")
	}

	// Wait for shutdown or timeout
	select {
	case <-shutdownCtx.Done():
		log.Warn().Msg("Shutdown timeout reached")
	case <-time.After(5 * time.Second):
		log.Info().Msg("Application shut down successfully")
	}

	fmt.Println("Application stopped")
}
