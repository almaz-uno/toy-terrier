package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/models"
)

// handleCallbackQuery processes callback queries from inline keyboards
func (h *Handlers) handleCallbackQuery(query *tgbotapi.CallbackQuery) error {
	// Update user activity
	if err := h.subscription.UpdateUserActivity(query.From.ID); err != nil {
		log.Error().Err(err).Msg("Failed to update user activity")
	}

	// Parse callback data
	data := query.Data
	parts := strings.Split(data, "_")

	if len(parts) < 1 {
		return h.answerCallbackQuery(query.ID, "❌ Неверная команда")
	}

	action := parts[0]

	switch action {
	case "subscribe":
		return h.handleSubscribeCallback(query, parts)
	case "unsubscribe":
		return h.handleUnsubscribeCallback(query, parts)
	case "category":
		return h.handleCategoryCallback(query, parts)
	case "settings":
		return h.handleSettingsCallback(query, parts)
	case "cancel":
		return h.handleCancelCallback(query)
	default:
		return h.answerCallbackQuery(query.ID, "❌ Неизвестная команда")
	}
}

// handleSubscribeCallback handles subscription callbacks
func (h *Handlers) handleSubscribeCallback(query *tgbotapi.CallbackQuery, parts []string) error {
	if len(parts) < 2 {
		return h.answerCallbackQuery(query.ID, "❌ Неверные параметры")
	}

	categoryID, err := strconv.Atoi(parts[1])
	if err != nil {
		return h.answerCallbackQuery(query.ID, "❌ Неверный ID категории")
	}

	userID := query.From.ID

	// Toggle subscription
	isSubscribed, err := h.subscription.IsUserSubscribed(userID, categoryID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check subscription")
		return h.answerCallbackQuery(query.ID, "❌ Ошибка проверки подписки")
	}

	if isSubscribed {
		err = h.subscription.Unsubscribe(userID, categoryID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unsubscribe")
			return h.answerCallbackQuery(query.ID, "❌ Ошибка отписки")
		}
		h.answerCallbackQuery(query.ID, "✅ Отписка выполнена")
	} else {
		err = h.subscription.Subscribe(userID, categoryID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to subscribe")
			return h.answerCallbackQuery(query.ID, "❌ Ошибка подписки")
		}
		h.answerCallbackQuery(query.ID, "✅ Подписка оформлена")
	}

	// Update keyboard
	categories, err := h.subscription.GetActiveCategories()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get categories")
		return nil
	}

	keyboard := h.buildCategoriesKeyboard(categories, userID, "subscribe")

	edit := tgbotapi.NewEditMessageReplyMarkup(query.Message.Chat.ID, query.Message.MessageID, keyboard)
	_, err = h.api.Send(edit)
	return err
}

// handleUnsubscribeCallback handles unsubscribe callbacks
func (h *Handlers) handleUnsubscribeCallback(query *tgbotapi.CallbackQuery, parts []string) error {
	if len(parts) < 2 {
		return h.answerCallbackQuery(query.ID, "❌ Неверные параметры")
	}

	if parts[1] == "all" {
		userID := query.From.ID
		err := h.subscription.UnsubscribeFromAll(userID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unsubscribe from all")
			return h.answerCallbackQuery(query.ID, "❌ Ошибка отписки")
		}

		// Update message
		text := "✅ <b>Вы отписались от всех уведомлений</b>\n\nИспользуйте /subscribe для повторной подписки."
		edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
		edit.ParseMode = tgbotapi.ModeHTML
		_, err = h.api.Send(edit)

		return h.answerCallbackQuery(query.ID, "✅ Отписка выполнена")
	}

	return h.answerCallbackQuery(query.ID, "❌ Неизвестная команда")
}

// handleCategoryCallback handles category-specific callbacks
func (h *Handlers) handleCategoryCallback(query *tgbotapi.CallbackQuery, parts []string) error {
	// Similar to subscribe callback but for category management
	return h.handleSubscribeCallback(query, parts[1:])
}

// handleSettingsCallback handles settings callbacks
func (h *Handlers) handleSettingsCallback(query *tgbotapi.CallbackQuery, parts []string) error {
	// TODO: Implement settings management
	return h.answerCallbackQuery(query.ID, "⚙️ Настройки пока не реализованы")
}

// handleCancelCallback handles cancel callbacks
func (h *Handlers) handleCancelCallback(query *tgbotapi.CallbackQuery) error {
	text := "❌ Операция отменена"
	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeHTML
	h.api.Send(edit)

	return h.answerCallbackQuery(query.ID, "Операция отменена")
}

// answerCallbackQuery answers a callback query
func (h *Handlers) answerCallbackQuery(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	callback.ShowAlert = false
	_, err := h.api.Request(callback)
	return err
}

// buildCategoriesKeyboard builds inline keyboard for categories
func (h *Handlers) buildCategoriesKeyboard(categories []*models.Category, userID int64, action string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// This is a placeholder - in real implementation, you'd need to define proper category structure
	// For now, create a simple keyboard
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ Hot Toys", fmt.Sprintf("%s_1", action)),
		tgbotapi.NewInlineKeyboardButtonData("❌ ThreeZero", fmt.Sprintf("%s_2", action)),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ Sideshow", fmt.Sprintf("%s_3", action)),
		tgbotapi.NewInlineKeyboardButtonData("❌ Facebook", fmt.Sprintf("%s_4", action)),
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✅ Готово", "cancel"),
	})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
