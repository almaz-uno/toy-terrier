package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

// handleAdminStats handles /admin_stats command
func (h *Handlers) handleAdminStats(message *tgbotapi.Message) error {
	if !h.isAdmin(message.From.ID) {
		return h.sendMessage(message.Chat.ID, "❌ У вас нет прав администратора")
	}

	stats, err := h.subscription.GetSystemStats()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get system stats")
		return h.sendMessage(message.Chat.ID, "❌ Ошибка при получении статистики")
	}

	var text strings.Builder
	text.WriteString("<b>📊 Статистика бота</b>\n\n")

	text.WriteString(fmt.Sprintf("👥 <b>Пользователи:</b> %d (%d активных)\n",
		stats.TotalUsers, stats.ActiveUsers))
	text.WriteString(fmt.Sprintf("📬 <b>Подписки:</b> %d\n", stats.TotalSubscriptions))
	text.WriteString(fmt.Sprintf("📝 <b>Постов обработано:</b> %d (сегодня: %d)\n",
		stats.TotalTopics, stats.TopicsToday))
	text.WriteString(fmt.Sprintf("📤 <b>Сообщений отправлено:</b> %d (сегодня: %d)\n",
		stats.TotalNotifications, stats.NotificationsToday))
	text.WriteString(fmt.Sprintf("❌ <b>Ошибок:</b> %d\n", stats.FailedNotifications))

	if stats.LastScrapedAt != nil {
		lastScrape := time.Since(*stats.LastScrapedAt)
		text.WriteString(fmt.Sprintf("🕐 <b>Последний парсинг:</b> %s назад\n",
			formatDuration(lastScrape)))
	}

	uptime := time.Since(stats.UptimeStart)
	text.WriteString(fmt.Sprintf("⏱ <b>Время работы:</b> %s", formatDuration(uptime)))

	return h.sendMessage(message.Chat.ID, text.String())
}

// handleAdminBroadcast handles /admin_broadcast command
func (h *Handlers) handleAdminBroadcast(message *tgbotapi.Message) error {
	if !h.isAdmin(message.From.ID) {
		return h.sendMessage(message.Chat.ID, "❌ У вас нет прав администратора")
	}

	// Extract message text after command
	args := strings.TrimSpace(strings.TrimPrefix(message.Text, "/admin_broadcast"))
	if args == "" {
		return h.sendMessage(message.Chat.ID,
			"❌ Укажите текст для рассылки:\n/admin_broadcast Ваше сообщение")
	}

	// Confirm broadcast
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Отправить", "broadcast_confirm_"+strconv.Itoa(message.MessageID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
		),
	)

	confirmText := fmt.Sprintf("<b>📢 Подтверждение рассылки</b>\n\n<b>Сообщение:</b>\n%s\n\n<b>Получатели:</b> Все активные пользователи\n\nПодтвердите отправку:", args)

	msg := tgbotapi.NewMessage(message.Chat.ID, confirmText)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err := h.api.Send(msg)
	return err
}

// handleAdminUsers handles /admin_users command
func (h *Handlers) handleAdminUsers(message *tgbotapi.Message) error {
	if !h.isAdmin(message.From.ID) {
		return h.sendMessage(message.Chat.ID, "❌ У вас нет прав администратора")
	}

	users, err := h.subscription.GetRecentUsers(20)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get recent users")
		return h.sendMessage(message.Chat.ID, "❌ Ошибка при получении списка пользователей")
	}

	var text strings.Builder
	text.WriteString("<b>👥 Последние пользователи</b>\n\n")

	for i, user := range users {
		username := "—"
		if user.Username != nil {
			username = "@" + *user.Username
		}

		status := "❌"
		if user.IsActive {
			status = "✅"
		}

		text.WriteString(fmt.Sprintf("%d. %s %s (%s)\n   ID: %d, активность: %s\n\n",
			i+1, status, user.FirstName, username, user.TelegramID,
			formatDuration(time.Since(user.LastActivity))))
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Полная статистика", "admin_full_stats"),
			tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "admin_users_refresh"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err = h.api.Send(msg)
	return err
}

// formatDuration formats duration in human readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0f сек", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f мин", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f ч", d.Hours())
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%d дн", days)
}

// Helper struct for stats (placeholder)
type SystemStats struct {
	TotalUsers          int
	ActiveUsers         int
	TotalSubscriptions  int
	TotalTopics         int
	TopicsToday         int
	TotalNotifications  int
	NotificationsToday  int
	FailedNotifications int
	LastScrapedAt       *time.Time
	UptimeStart         time.Time
}

// Helper struct for user info (placeholder)
type UserInfo struct {
	TelegramID   int64
	Username     *string
	FirstName    string
	IsActive     bool
	LastActivity time.Time
}
