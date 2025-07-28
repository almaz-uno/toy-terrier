package handlers

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/config"
	"toy-terrier-telegram/internal/services/notification"
	"toy-terrier-telegram/internal/services/subscription"
)

// Handlers manages all bot command handlers
type Handlers struct {
	config       *config.Config
	db           *sql.DB
	api          *tgbotapi.BotAPI
	subscription *subscription.Manager
	notification *notification.Service
}

// New creates a new handlers instance
func New(cfg *config.Config, db *sql.DB, api *tgbotapi.BotAPI,
	sub *subscription.Manager, notif *notification.Service,
) *Handlers {
	return &Handlers{
		config:       cfg,
		db:           db,
		api:          api,
		subscription: sub,
		notification: notif,
	}
}

// HandleUpdate routes updates to appropriate handlers
func (h *Handlers) HandleUpdate(update tgbotapi.Update) error {
	// Handle messages
	if update.Message != nil {
		return h.handleMessage(update.Message)
	}

	// Handle callback queries
	if update.CallbackQuery != nil {
		return h.handleCallbackQuery(update.CallbackQuery)
	}

	return nil
}

// handleMessage processes incoming messages
func (h *Handlers) handleMessage(message *tgbotapi.Message) error {
	// Ignore non-text messages for now
	if !message.IsCommand() {
		return nil
	}

	// Update user activity
	if err := h.subscription.UpdateUserActivity(message.From.ID); err != nil {
		log.Error().Err(err).Msg("Failed to update user activity")
	}

	// Route command
	switch message.Command() {
	case "start":
		return h.handleStart(message)
	case "help":
		return h.handleHelp(message)
	case "subscribe":
		return h.handleSubscribe(message)
	case "unsubscribe":
		return h.handleUnsubscribe(message)
	case "categories":
		return h.handleCategories(message)
	case "settings":
		return h.handleSettings(message)
	case "status":
		return h.handleStatus(message)
	case "admin_stats":
		return h.handleAdminStats(message)
	case "admin_broadcast":
		return h.handleAdminBroadcast(message)
	case "admin_users":
		return h.handleAdminUsers(message)
	default:
		return h.handleUnknownCommand(message)
	}
}

// handleStart handles /start command
func (h *Handlers) handleStart(message *tgbotapi.Message) error {
	user := message.From

	// Create or update user
	if err := h.subscription.CreateOrUpdateUser(user); err != nil {
		log.Error().Err(err).Msg("Failed to create/update user")
		return h.sendMessage(message.Chat.ID, "❌ Произошла ошибка. Попробуйте позже.")
	}

	welcomeText := `🎭 <b>Добро пожаловать в бот мониторинга фигурок!</b>

Я буду уведомлять вас о новых постах с:
• BBS BBICN Forum - форум коллекционеров фигурок
• Facebook Hot Toys - официальная страница Hot Toys

<b>Команды:</b>
/subscribe - подписка на уведомления
/categories - управление категориями
/settings - настройки
/status - статус подписки
/help - справка

Используйте /subscribe для подписки на уведомления`

	return h.sendMessage(message.Chat.ID, welcomeText)
}

// handleHelp handles /help command
func (h *Handlers) handleHelp(message *tgbotapi.Message) error {
	helpText := `<b>🤖 Справка по командам</b>

<b>Основные команды:</b>
/start - начать работу с ботом
/subscribe - подписаться на уведомления
/unsubscribe - отписаться от уведомлений
/categories - управление подписками по категориям
/settings - персональные настройки
/status - текущий статус подписки
/help - эта справка

<b>Что умеет бот:</b>
• Мониторит форум BBS BBICN (Hot Toys, ThreeZero, Sideshow)
• Отслеживает посты Facebook Hot Toys
• Отправляет уведомления о новых темах и фигурках
• Позволяет настроить категории интересов

<b>Поддержка:</b>
По всем вопросам обращайтесь к администратору.`

	return h.sendMessage(message.Chat.ID, helpText)
}

// handleSubscribe handles /subscribe command
func (h *Handlers) handleSubscribe(message *tgbotapi.Message) error {
	categories, err := h.subscription.GetActiveCategories()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get categories")
		return h.sendMessage(message.Chat.ID, "❌ Ошибка при получении категорий")
	}

	if len(categories) == 0 {
		return h.sendMessage(message.Chat.ID, "❌ Нет доступных категорий для подписки")
	}

	keyboard := h.buildCategoriesKeyboard(categories, message.From.ID, "subscribe")

	text := `<b>📝 Подписка на обновления</b>

Выберите категории для подписки на новые посты и обновления.

Нажмите на кнопки ниже, чтобы подписаться на интересующие вас категории.`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err = h.api.Send(msg)
	return err
}

// handleUnsubscribe handles /unsubscribe command
func (h *Handlers) handleUnsubscribe(message *tgbotapi.Message) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Да, отписаться от всех", "unsubscribe_all"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "cancel"),
		),
	)

	text := `<b>🚫 Отписка от уведомлений</b>

Вы уверены, что хотите отписаться от всех уведомлений?

Это действие можно будет отменить с помощью команды /subscribe.`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err := h.api.Send(msg)
	return err
}

// handleCategories handles /categories command
func (h *Handlers) handleCategories(message *tgbotapi.Message) error {
	subscriptions, err := h.subscription.GetUserSubscriptions(message.From.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user subscriptions")
		return h.sendMessage(message.Chat.ID, "❌ Ошибка при получении подписок")
	}

	if len(subscriptions) == 0 {
		return h.sendMessage(message.Chat.ID, "У вас нет активных подписок. Используйте /subscribe для подписки.")
	}

	text := "<b>📂 Ваши подписки:</b>\n\n"
	for _, sub := range subscriptions {
		status := "✅"
		if !sub.IsActive {
			status = "❌"
		}
		text += fmt.Sprintf("%s <b>%s</b>\n", status, sub.Name)
	}

	text += "\nИспользуйте /subscribe для управления подписками."

	return h.sendMessage(message.Chat.ID, text)
}

// handleSettings handles /settings command
func (h *Handlers) handleSettings(message *tgbotapi.Message) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 Управление подписками", "manage_subscriptions"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔔 Настройки уведомлений", "notification_settings"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Закрыть", "cancel"),
		),
	)

	text := `<b>⚙️ Настройки</b>

Выберите раздел для настройки:`

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = keyboard

	_, err := h.api.Send(msg)
	return err
}

// handleStatus handles /status command
func (h *Handlers) handleStatus(message *tgbotapi.Message) error {
	subscriptions, err := h.subscription.GetUserSubscriptions(message.From.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user subscriptions")
		return h.sendMessage(message.Chat.ID, "❌ Ошибка при получении подписок")
	}

	activeCount := 0
	for _, sub := range subscriptions {
		if sub.IsActive {
			activeCount++
		}
	}

	username := message.From.UserName
	if username == "" {
		username = message.From.FirstName
	}

	text := fmt.Sprintf(`<b>📊 Ваш статус</b>

<b>Пользователь:</b> %s
<b>Активных подписок:</b> %d из %d

<b>Последние уведомления:</b>
Сегодня получено уведомлений: 0`,
		username, activeCount, len(subscriptions))

	return h.sendMessage(message.Chat.ID, text)
}

// handleUnknownCommand handles unknown commands
func (h *Handlers) handleUnknownCommand(message *tgbotapi.Message) error {
	text := `❓ Неизвестная команда.

Используйте /help для просмотра доступных команд.`

	return h.sendMessage(message.Chat.ID, text)
}

// sendMessage sends a message to chat
func (h *Handlers) sendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := h.api.Send(msg)
	return err
}

// isAdmin checks if user is admin
func (h *Handlers) isAdmin(userID int64) bool {
	for _, adminID := range h.config.Telegram.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}
