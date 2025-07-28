# Toy Terrier Telegram Bot

Telegram бот для мониторинга форума BBS BBICN и Facebook Hot Toys с уведомлениями о новых постах.

## Возможности

- 🤖 Telegram бот с polling архитектурой
- 📊 Мониторинг форума BBS BBICN и Facebook Hot Toys
- 🔔 Уведомления о новых постах
- 👥 Система подписок по категориям
- 📈 HTTP API для администрирования
- 🗄️ PostgreSQL база данных
- 🐳 Docker поддержка

## Мониторируемые источники

### 🏛️ BBS BBICN Forum (https://bbs.bbicn.com/)
Специализированный форум коллекционеров фигурок:
- **Hot Toys** - Официальные анонсы и обсуждения фигурок Hot Toys
- **ThreeZero** - Новинки от производителя ThreeZero  
- **Sideshow** - Коллекционные фигурки Sideshow Collectibles

### 📘 Facebook Hot Toys (https://www.facebook.com/hottoys)
Официальная страница производителя Hot Toys:
- Официальные анонсы новых фигурок
- Превью и тизеры предстоящих релизов
- Высококачественные фотографии продукции

## Архитектура

Проект построен на:
- **Go 1.24+** - основной язык разработки
- **PostgreSQL 15+** - база данных
- **Telegram Bot API** - интеграция с Telegram (polling mode)
- **Echo v4** - HTTP веб-фреймворк для админки
- **golang-migrate** - миграции базы данных
- **zerolog** - структурированное логирование
- **goquery** - парсинг HTML

## Структура проекта

```
.
├── cmd/toy-terrier-bot/        # Точка входа приложения
├── internal/
│   ├── config/                # Конфигурация
│   ├── database/              # Подключение к БД и миграции
│   ├── models/                # Модели данных
│   ├── bot/                   # Telegram бот
│   │   ├── handlers/          # Обработчики команд
│   │   └── polling/           # Polling сервис
│   ├── services/              # Бизнес-логика
│   │   ├── notification/      # Сервис уведомлений
│   │   ├── scraper/          # Парсинг форумов
│   │   └── subscription/     # Управление подписками
│   └── server/               # HTTP сервер
├── docs/                     # Документация
└── Dockerfile               # Docker образ
```

## Установка и запуск

### Требования

- Go 1.24+
- PostgreSQL 15+
- Telegram Bot Token

### Настройка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/almaz-uno/toy-terrier.git
cd toy-terrier
```

2. Создайте Telegram бота через [@BotFather](https://t.me/botfather) и получите токен

3. (Опционально) Настройте Facebook App для мониторинга Hot Toys:
   - Создайте Facebook App в [Facebook Developers](https://developers.facebook.com/)
   - Получите App ID, App Secret и Access Token
   - Укажите Page ID для страницы Hot Toys

4. Скопируйте пример конфигурации:
```bash
cp config.example.yaml config.yaml
```

5. Отредактируйте `config.yaml` и укажите:
   - `telegram.bot_token`: Токен вашего Telegram бота
   - `database.*`: Настройки подключения к PostgreSQL  
   - `security.admin_ids`: Ваш Telegram User ID для администрирования
   - `facebook.*`: (Опционально) Настройки Facebook API для Hot Toys

6. Настройте базу данных PostgreSQL:
```bash
createdb terrier_telegram
```

### Конфигурация

Основные параметры в `config.yaml`:

```yaml
telegram:
  bot_token: "YOUR_BOT_TOKEN"
  admin_ids: [123456789]

database:
  host: "localhost"
  user: "postgres"
  password: "password"
  name: "toy_terrier"

scraper:
  scrape_interval: "5m"
  
notification:
  queue_check_interval: "10s"
```

### Запуск

```bash
# Разработка
go run ./cmd/toy-terrier-bot

# Продакшен
go build -o toy-terrier-bot ./cmd/toy-terrier-bot
./toy-terrier-bot
```

### Docker

```bash
docker build -t toy-terrier-bot .
docker run -d --name toy-terrier \
  -v $(pwd)/config.yaml:/app/config.yaml \
  toy-terrier-bot
```

## API

HTTP сервер запускается на порту 8080 (настраивается в config.yaml).

### Эндпоинты

- `GET /health` - проверка здоровья
- `GET /api/stats` - общая статистика
- `GET /api/users` - список пользователей
- `GET /api/subscriptions` - подписки
- `POST /api/notifications/broadcast` - рассылка сообщений

## Команды бота

- `/start` - начать работу с ботом
- `/help` - справка по командам
- `/subscribe` - подписаться на уведомления
- `/unsubscribe` - отписаться от уведомлений
- `/status` - текущий статус подписки

### Админские команды

- `/admin_stats` - статистика системы
- `/admin_broadcast` - рассылка сообщений
- `/admin_users` - управление пользователями

## База данных

Схема базы данных включает:

- `users` - пользователи бота
- `categories` - категории форумов
- `subscriptions` - подписки пользователей
- `forum_posts` - спарсенные посты
- `notifications` - очередь уведомлений

## Разработка

### Тестирование

```bash
go test ./...
```

### Линтинг

```bash
golangci-lint run
```

### Миграции

Создание новой миграции:
```bash
migrate create -ext sql -dir internal/database/migrations -seq migration_name
```

## Мониторинг

- Логи в JSON формате (настраивается)
- Метрики через HTTP API
- Health check эндпоинт

## Лицензия

MIT License

## Поддержка

По вопросам обращайтесь к администратору бота или создавайте issue в репозитории.
