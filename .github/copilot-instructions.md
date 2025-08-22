# GitHub Copilot Instructions for Toy Terrier Bot

## Project Overview
Toy Terrier is a Telegram bot written in Go that provides RSS feed monitoring and notification services. The bot scrapes forum posts and sends notifications to subscribed users.

## Project Structure
- `cmd/toy-terrier-bot/` - Main application entry point
- `internal/` - Core business logic
  - `bot/` - Telegram bot implementation
  - `config/` - Configuration management
  - `database/` - Database layer with migrations and queries
  - `models/` - Data models
  - `server/` - HTTP server
  - `services/` - Business services (notification, scraper, subscription)

## Code Style and Conventions

### Go Conventions
- Follow standard Go naming conventions (PascalCase for exported, camelCase for unexported)
- Use gofmt for formatting
- Write clear, descriptive variable and function names
- Add comments for exported functions and complex logic
- Use context.Context for cancellation and timeouts
- Handle errors explicitly, don't ignore them
- Documentation should be in AsciiDoc format, unless otherwise explicitly specified. Use .<title> as a title for ordered or unordered lists.

### Database
- Use SQLC for type-safe SQL queries
- Store SQL queries in `internal/database/queries/`
- Migrations are in `internal/database/migrations/`
- Follow naming pattern: `XXX_description.up.sql` and `XXX_description.down.sql`

### Project-Specific Patterns
- Use manual dependency injection (no DI framework like Wire/Dig/Fx)
- Pass dependencies explicitly through constructors
- Implement interfaces for testability
- Store configuration in `config.yaml`
- Use structured logging with zerolog
- Handle Telegram API rate limits gracefully

## Dependencies
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- Database migrations and queries managed with SQLC
- Configuration with YAML
- Standard Go libraries for HTTP, context, etc.

## Development Workflow
- Use `make lint` for linting
- Use `make test` for running tests
- Use `make build` for building the application
- Docker support available via `Dockerfile` and `docker-compose.yml`

## When Writing Code
1. Always check if similar functionality already exists
2. Follow the existing patterns in the codebase
3. Add appropriate error handling
4. Consider adding tests for new functionality
5. Update documentation if adding new features
6. Use the established project structure

## Telegram Bot Specifics
- Handle different message types appropriately
- Implement proper callback handling
- Use inline keyboards for user interactions
- Handle admin commands separately from user commands
- Implement proper user state management

## Database Considerations
- Always use transactions for multi-step operations
- Use prepared statements via SQLC generated code
- Handle database connection errors gracefully
- Consider data consistency and foreign key relationships

## Security
- Validate all user inputs
- Use environment variables for sensitive data
- Implement proper authentication for admin functions
- Sanitize data before database operations
