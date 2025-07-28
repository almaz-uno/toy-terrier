package server

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/config"
	"toy-terrier-telegram/internal/services/notification"
	"toy-terrier-telegram/internal/services/scraper"
)

// Server represents HTTP server
type Server struct {
	config              *config.Config
	db                  *sql.DB
	notificationService *notification.Service
	scraperService      *scraper.Service
	echo                *echo.Echo
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *sql.DB,
	notificationService *notification.Service, scraperService *scraper.Service,
) *Server {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	server := &Server{
		config:              cfg,
		db:                  db,
		notificationService: notificationService,
		scraperService:      scraperService,
		echo:                e,
	}

	server.setupRoutes()
	return server
}

// setupRoutes sets up HTTP routes
func (s *Server) setupRoutes() {
	// Health check
	s.echo.GET("/health", s.healthCheck)

	// API routes
	api := s.echo.Group("/api")

	// Statistics
	api.GET("/stats", s.getStats)

	// Users
	api.GET("/users", s.getUsers)
	api.GET("/users/:id", s.getUser)

	// Subscriptions
	api.GET("/subscriptions", s.getSubscriptions)

	// Notifications
	api.GET("/notifications/queue", s.getNotificationQueue)
	api.POST("/notifications/broadcast", s.broadcastMessage)

	// Scraping
	api.GET("/scraping/metrics", s.getScrapingMetrics)
	api.GET("/posts", s.getPosts)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + strconv.Itoa(s.config.Server.Port)
	log.Info().Str("addr", addr).Msg("Starting HTTP server")
	return s.echo.Start(addr)
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	log.Info().Msg("Stopping HTTP server")
	return s.echo.Close()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.Stop()
}

// healthCheck handles health check requests
func (s *Server) healthCheck(c echo.Context) error {
	// Check database connection
	if err := s.db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z", // Replace with actual timestamp
	})
}

// getStats handles statistics requests
func (s *Server) getStats(c echo.Context) error {
	stats := make(map[string]interface{})

	// Get user stats
	var totalUsers, activeUsers int
	err := s.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM users
	`).Scan(&totalUsers, &activeUsers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user stats")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user stats",
		})
	}

	stats["users"] = map[string]int{
		"total":  totalUsers,
		"active": activeUsers,
	}

	// Get subscription stats
	var totalSubscriptions, activeSubscriptions int
	err = s.db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE is_active = true) as active
		FROM subscriptions
	`).Scan(&totalSubscriptions, &activeSubscriptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription stats")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get subscription stats",
		})
	}

	stats["subscriptions"] = map[string]int{
		"total":  totalSubscriptions,
		"active": activeSubscriptions,
	}

	// Get notification queue metrics
	queueMetrics, err := s.notificationService.GetQueueMetrics()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get queue metrics")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get queue metrics",
		})
	}

	stats["notifications"] = queueMetrics

	// Get scraping metrics
	scrapingMetrics, err := s.scraperService.GetScrapingMetrics()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get scraping metrics")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get scraping metrics",
		})
	}

	stats["scraping"] = scrapingMetrics

	return c.JSON(http.StatusOK, stats)
}

// getUsers handles user list requests
func (s *Server) getUsers(c echo.Context) error {
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := `
		SELECT id, telegram_id, username, first_name, last_name, is_active, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get users",
		})
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var telegramID int64
		var username, firstName, lastName string
		var isActive bool
		var createdAt string

		err := rows.Scan(&id, &telegramID, &username, &firstName, &lastName, &isActive, &createdAt)
		if err != nil {
			continue
		}

		user := map[string]interface{}{
			"id":          id,
			"telegram_id": telegramID,
			"username":    username,
			"first_name":  firstName,
			"last_name":   lastName,
			"is_active":   isActive,
			"created_at":  createdAt,
		}

		users = append(users, user)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	})
}

// getUser handles single user requests
func (s *Server) getUser(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid user ID",
		})
	}

	query := `
		SELECT id, telegram_id, username, first_name, last_name, is_active, created_at
		FROM users
		WHERE id = $1
	`

	var id int
	var telegramID int64
	var username, firstName, lastName string
	var isActive bool
	var createdAt string

	err = s.db.QueryRow(query, userID).Scan(&id, &telegramID, &username, &firstName, &lastName, &isActive, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		log.Error().Err(err).Msg("Failed to get user")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user",
		})
	}

	user := map[string]interface{}{
		"id":          id,
		"telegram_id": telegramID,
		"username":    username,
		"first_name":  firstName,
		"last_name":   lastName,
		"is_active":   isActive,
		"created_at":  createdAt,
	}

	// Get user subscriptions
	subsQuery := `
		SELECT s.id, c.name, s.is_active, s.created_at
		FROM subscriptions s
		JOIN categories c ON s.category_id = c.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`

	rows, err := s.db.Query(subsQuery, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user subscriptions")
	} else {
		defer rows.Close()
		var subscriptions []map[string]interface{}

		for rows.Next() {
			var subID int
			var categoryName string
			var subIsActive bool
			var subCreatedAt string

			if err := rows.Scan(&subID, &categoryName, &subIsActive, &subCreatedAt); err == nil {
				subscription := map[string]interface{}{
					"id":            subID,
					"category_name": categoryName,
					"is_active":     subIsActive,
					"created_at":    subCreatedAt,
				}
				subscriptions = append(subscriptions, subscription)
			}
		}

		user["subscriptions"] = subscriptions
	}

	return c.JSON(http.StatusOK, user)
}

// getSubscriptions handles subscription list requests
func (s *Server) getSubscriptions(c echo.Context) error {
	query := `
		SELECT s.id, u.username, c.name, s.is_active, s.created_at
		FROM subscriptions s
		JOIN users u ON s.user_id = u.id
		JOIN categories c ON s.category_id = c.id
		ORDER BY s.created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Query(query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscriptions")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get subscriptions",
		})
	}
	defer rows.Close()

	var subscriptions []map[string]interface{}
	for rows.Next() {
		var id int
		var username, categoryName string
		var isActive bool
		var createdAt string

		err := rows.Scan(&id, &username, &categoryName, &isActive, &createdAt)
		if err != nil {
			continue
		}

		subscription := map[string]interface{}{
			"id":            id,
			"username":      username,
			"category_name": categoryName,
			"is_active":     isActive,
			"created_at":    createdAt,
		}

		subscriptions = append(subscriptions, subscription)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
	})
}

// getNotificationQueue handles notification queue requests
func (s *Server) getNotificationQueue(c echo.Context) error {
	metrics, err := s.notificationService.GetQueueMetrics()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get notification queue")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get notification queue",
		})
	}

	return c.JSON(http.StatusOK, metrics)
}

// broadcastMessage handles broadcast message requests
func (s *Server) broadcastMessage(c echo.Context) error {
	var request struct {
		Message string `json:"message"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if request.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Message is required",
		})
	}

	err := s.notificationService.BroadcastToAllUsers(request.Message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to broadcast message")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to broadcast message",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "Message queued for broadcast",
	})
}

// getScrapingMetrics handles scraping metrics requests
func (s *Server) getScrapingMetrics(c echo.Context) error {
	metrics, err := s.scraperService.GetScrapingMetrics()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get scraping metrics")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get scraping metrics",
		})
	}

	return c.JSON(http.StatusOK, metrics)
}

// getPosts handles forum posts requests
func (s *Server) getPosts(c echo.Context) error {
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	query := `
		SELECT fp.id, fp.title, fp.url, fp.author, c.name, fp.created_at, fp.scraped_at
		FROM forum_posts fp
		JOIN categories c ON fp.category_id = c.id
		ORDER BY fp.scraped_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get posts")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get posts",
		})
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var id int
		var title, url, author, categoryName string
		var createdAt, scrapedAt string

		err := rows.Scan(&id, &title, &url, &author, &categoryName, &createdAt, &scrapedAt)
		if err != nil {
			continue
		}

		post := map[string]interface{}{
			"id":            id,
			"title":         title,
			"url":           url,
			"author":        author,
			"category_name": categoryName,
			"created_at":    createdAt,
			"scraped_at":    scrapedAt,
		}

		posts = append(posts, post)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"posts":  posts,
		"limit":  limit,
		"offset": offset,
	})
}
