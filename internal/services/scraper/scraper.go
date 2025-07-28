package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"

	"toy-terrier-telegram/internal/config"
	"toy-terrier-telegram/internal/models"
	"toy-terrier-telegram/internal/services/notification"
)

// Service handles forum scraping operations
type Service struct {
	config              *config.Config
	db                  *sql.DB
	notificationService *notification.Service
	client              *http.Client
}

// NewService creates a new scraper service
func NewService(cfg *config.Config, db *sql.DB, notificationService *notification.Service) *Service {
	client := &http.Client{
		Timeout: time.Duration(cfg.Scraper.TimeoutSeconds) * time.Second,
	}

	return &Service{
		config:              cfg,
		db:                  db,
		notificationService: notificationService,
		client:              client,
	}
}

// Start starts the scraper service
func (s *Service) Start(ctx context.Context) error {
	log.Info().Msg("Starting scraper service")

	ticker := time.NewTicker(s.config.Scraper.ScrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Scraper service stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := s.scrapeAllForums(); err != nil {
				log.Error().Err(err).Msg("Failed to scrape forums")
			}
		}
	}
}

// scrapeAllForums scrapes all active forums
func (s *Service) scrapeAllForums() error {
	// Get all active categories
	query := `
		SELECT id, name, slug, source_type, source_url, is_active, created_at, updated_at
		FROM categories
		WHERE is_active = true
		ORDER BY id
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category

	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Slug,
			&category.SourceType, &category.SourceURL, &category.IsActive,
			&category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan category")
			continue
		}
		categories = append(categories, &category)
	}

	// Scrape each category
	for _, category := range categories {
		if err := s.scrapeCategory(category); err != nil {
			log.Error().
				Err(err).
				Int("category_id", category.ID).
				Str("source_type", category.SourceType).
				Msg("Failed to scrape category")
		}

		// Rate limiting between requests
		time.Sleep(s.config.Scraper.RequestDelay)
	}

	return nil
}

// scrapeCategory scrapes a specific category
func (s *Service) scrapeCategory(category *models.Category) error {
	log.Debug().
		Str("source_url", category.SourceURL).
		Int("category_id", category.ID).
		Str("source_type", category.SourceType).
		Msg("Scraping category")

	var newPosts []*models.ForumPost
	var err error

	// Handle different source types
	switch category.SourceType {
	case "forum": // BBS BBICN Forum
		newPosts, err = s.scrapeBBSForumPosts(category)
	case "facebook": // Facebook Hot Toys
		newPosts, err = s.scrapeFacebookPosts(category)
	default:
		return fmt.Errorf("unsupported source type: %s", category.SourceType)
	}

	if err != nil {
		return fmt.Errorf("failed to scrape posts: %w", err)
	}

	// Process new posts
	for _, post := range newPosts {
		if err := s.processNewPost(post); err != nil {
			log.Error().
				Err(err).
				Str("title", post.Title).
				Msg("Failed to process new post")
		}
	}

	log.Info().
		Int("category_id", category.ID).
		Int("new_posts", len(newPosts)).
		Msg("Scraped category")

	return nil
}

// scrapeBBSForumPosts scrapes posts from BBS BBICN forum
func (s *Service) scrapeBBSForumPosts(category *models.Category) ([]*models.ForumPost, error) {
	req, err := http.NewRequest("GET", category.SourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.Scraper.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var posts []*models.ForumPost

	// Parse BBS BBICN forum structure
	doc.Find("tbody[id^='normalthread_'] tr").Each(func(i int, tr *goquery.Selection) {
		// Skip sticky posts and announcements
		if tr.HasClass("bg_pin") || tr.HasClass("bg_top") {
			return
		}

		// Find title link
		titleCell := tr.Find("td.icn").Next() // Thread title is usually in the cell after icon
		if titleCell.Length() == 0 {
			titleCell = tr.Find("td.subject")
		}

		link := titleCell.Find("a[href*='thread-']").First()
		title := strings.TrimSpace(link.Text())
		href, exists := link.Attr("href")
		if !exists || title == "" {
			return
		}

		// Build full URL if relative
		if !strings.HasPrefix(href, "http") {
			href = "https://bbs.bbicn.com/" + strings.TrimPrefix(href, "/")
		}

		// Get author
		authorCell := tr.Find("td.author").Find("a").First()
		author := strings.TrimSpace(authorCell.Text())

		// Get creation time
		timeCell := tr.Find("td.lastpost")
		timeStr := strings.TrimSpace(timeCell.Text())

		post := &models.ForumPost{
			CategoryID: category.ID,
			Title:      title,
			URL:        href,
			Author:     author,
			CreatedAt:  s.parseDateTime(timeStr),
			ScrapedAt:  time.Now(),
		}

		posts = append(posts, post)
	})

	return posts, nil
}

// scrapeFacebookPosts scrapes posts from Facebook Hot Toys page
func (s *Service) scrapeFacebookPosts(category *models.Category) ([]*models.ForumPost, error) {
	// For Facebook, we would use Graph API instead of web scraping
	// This is a placeholder implementation
	log.Info().
		Str("category", category.Name).
		Msg("Facebook scraping not implemented yet - would use Graph API")

	// TODO: Implement Facebook Graph API integration
	// For now, return empty slice
	return []*models.ForumPost{}, nil
}

// parseDateTime parses forum date/time strings
func (s *Service) parseDateTime(timeStr string) time.Time {
	// Clean up the time string
	timeStr = strings.TrimSpace(timeStr)

	// Try different formats including BBS BBICN formats
	formats := []string{
		"2006-1-2 15:04",      // BBS BBICN format
		"2006-01-02 15:04:05", // Standard SQL format
		"02.01.2006 15:04",    // European format
		"15:04 02.01.2006",    // Time first format
		"Jan 2, 2006 15:04",   // English format
		"2006-1-2",            // Date only BBS format
		"2006-01-02",          // Date only standard
		"昨天 15:04",            // Chinese "yesterday"
		"今天 15:04",            // Chinese "today"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	// Handle relative dates in Chinese
	now := time.Now()
	if strings.Contains(timeStr, "昨天") { // Yesterday
		if timeOnly := strings.TrimPrefix(timeStr, "昨天 "); timeOnly != timeStr {
			if t, err := time.Parse("15:04", timeOnly); err == nil {
				yesterday := now.AddDate(0, 0, -1)
				return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(),
					t.Hour(), t.Minute(), 0, 0, now.Location())
			}
		}
		return now.AddDate(0, 0, -1)
	}

	if strings.Contains(timeStr, "今天") { // Today
		if timeOnly := strings.TrimPrefix(timeStr, "今天 "); timeOnly != timeStr {
			if t, err := time.Parse("15:04", timeOnly); err == nil {
				return time.Date(now.Year(), now.Month(), now.Day(),
					t.Hour(), t.Minute(), 0, 0, now.Location())
			}
		}
		return now
	}

	// If parsing fails, return current time
	log.Warn().Str("time_str", timeStr).Msg("Failed to parse time")
	return time.Now()
}

// processNewPost processes a new forum post
func (s *Service) processNewPost(post *models.ForumPost) error {
	// Check if post already exists
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM forum_posts WHERE url = $1)
	`, post.URL).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if post exists: %w", err)
	}

	if exists {
		return nil // Post already exists
	}

	// Insert new post
	query := `
		INSERT INTO forum_posts (category_id, title, url, author, content, created_at, scraped_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var postID int
	err = s.db.QueryRow(query, post.CategoryID, post.Title, post.URL,
		post.Author, post.Content, post.CreatedAt, post.ScrapedAt).Scan(&postID)
	if err != nil {
		return fmt.Errorf("failed to insert post: %w", err)
	}

	post.ID = postID

	log.Info().
		Int("post_id", postID).
		Int("category_id", post.CategoryID).
		Str("title", post.Title).
		Msg("New post found")

	// Send notifications to subscribers
	messageText := fmt.Sprintf(
		"<b>Новый пост в категории</b>\n\n"+
			"📝 <b>%s</b>\n"+
			"👤 Автор: %s\n"+
			"🔗 <a href=\"%s\">Читать полностью</a>",
		post.Title, post.Author, post.URL)

	if err := s.notificationService.NotifySubscribers(post.CategoryID, postID, messageText); err != nil {
		log.Error().
			Err(err).
			Int("post_id", postID).
			Msg("Failed to notify subscribers")
	}

	return nil
}

// GetScrapingMetrics returns scraping metrics
func (s *Service) GetScrapingMetrics() (*models.ScrapingMetrics, error) {
	metrics := &models.ScrapingMetrics{}

	// Get total posts count
	err := s.db.QueryRow(`SELECT COUNT(*) FROM forum_posts`).Scan(&metrics.TotalPosts)
	if err != nil {
		return nil, fmt.Errorf("failed to get total posts: %w", err)
	}

	// Get posts scraped in last 24 hours
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM forum_posts
		WHERE scraped_at > NOW() - INTERVAL '24 hours'
	`).Scan(&metrics.PostsLast24h)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts last 24h: %w", err)
	}

	// Get last scraping time
	err = s.db.QueryRow(`
		SELECT scraped_at FROM forum_posts
		ORDER BY scraped_at DESC
		LIMIT 1
	`).Scan(&metrics.LastScrapingAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get last scraping time: %w", err)
	}

	// Get posts by category
	query := `
		SELECT c.name, COUNT(fp.id)
		FROM categories c
		LEFT JOIN forum_posts fp ON c.id = fp.category_id
		GROUP BY c.id, c.name
		ORDER BY c.name
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by category: %w", err)
	}
	defer rows.Close()

	metrics.PostsByCategory = make(map[string]int)
	for rows.Next() {
		var categoryName string
		var count int
		if err := rows.Scan(&categoryName, &count); err != nil {
			continue
		}
		metrics.PostsByCategory[categoryName] = count
	}

	return metrics, nil
}
