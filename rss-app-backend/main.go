package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mmcdole/gofeed"
	"github.com/microcosm-cc/bluemonday"
	_ "github.com/jackc/pgx/v5/stdlib"
)


var DB *sql.DB

// Data Structures

type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Link    string `json:"link"`
	PublishedAt   string `json:"publishedAt"`
	FeedSourceURL string `json:"feedSourceUrl"`
}

type AddFeedRequest struct {
	URL string `json:"url"`
}

type RemoveFeedRequest struct {
	URL string `json:"url"`
}

type ArticlesByIdsRequest struct {
	IDs []int `json:"ids"`
}


func main() {
	// Connect to the database
	connStr := "user=kava password=password host=localhost port=5432 dbname=box"

	var err error
	DB, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer DB.Close()

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}
	log.Println("Successfully connected to the database.")

	// Set up the router
	r := mux.NewRouter()
	
	// API routes are grouped under /api
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/feeds", getFeedsHandler).Methods("GET")
	api.HandleFunc("/feeds", addFeedHandler).Methods("POST")
	api.HandleFunc("/feeds", removeFeedHandler).Methods("DELETE")
	api.HandleFunc("/articles", getArticlesHandler).Methods("GET")
	api.HandleFunc("/articles/by-ids", getArticlesByIdsHandler).Methods("POST")
	api.HandleFunc("/articles/update", updateArticlesHandler).Methods("POST")
	
	// Set up CORS for development
	// This allows the React dev server (on port 5173) to communicate with the Go backend (on port 8080)
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:5173"}), 
		handlers.AllowedMethods([]string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
	)

	// Start the server
	port := "8080"
	log.Printf("Starting API server on port %s", port)
	if err := http.ListenAndServe(":"+port, corsHandler(r)); err != nil {
		log.Fatal(err)
	}
}


// API Handlers

func getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT url FROM feeds ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer rows.Close()

	// Initialize as a non-nil empty slice
	feeds := make([]string, 0)

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			log.Printf("DB scan error: %v", err)
			return
		}
		feeds = append(feeds, url)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeds)
}

func addFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req AddFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	_, err := DB.Exec("INSERT INTO feeds (url) VALUES ($1)", req.URL)
	if err != nil {
		http.Error(w, "Failed to add feed", http.StatusInternalServerError)
		log.Printf("DB insert error: %v", err)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func removeFeedHandler(w http.ResponseWriter, r *http.Request) {
	var req RemoveFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := DB.Exec("DELETE FROM feeds WHERE url = $1", req.URL)
	if err != nil {
		http.Error(w, "Failed to remove feed", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getArticlesHandler(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT a.id, a.title, a.content, a.link, a.published_at, f.url
		FROM articles a
		LEFT JOIN feeds f ON a.feed_id = f.id
		ORDER BY a.published_at DESC NULLS LAST`

	rows, err := DB.Query(query)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	articles := make([]Article, 0)
	for rows.Next() {
		var a Article
		var publishedAt sql.NullTime
		var feedURL sql.NullString

		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Link, &publishedAt, &feedURL); err != nil {
			http.Error(w, "Failed to scan article", http.StatusInternalServerError)
			return
		}

		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if feedURL.Valid {
			a.FeedSourceURL = feedURL.String
		}

		articles = append(articles, a)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}

func getArticlesByIdsHandler(w http.ResponseWriter, r *http.Request) {
	var req ArticlesByIdsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Article{}) // Return empty array if no IDs provided
		return
	}

	// pgx uses ANY($1) to handle IN clauses with slices
	query := `
		SELECT a.id, a.title, a.content, a.link, a.published_at, f.url
		FROM articles a
		LEFT JOIN feeds f ON a.feed_id = f.id
		WHERE a.id = ANY($1)
		ORDER BY a.published_at DESC NULLS LAST`

	rows, err := DB.Query(query, req.IDs)
	
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		log.Printf("getArticlesByIdsHandler query error: %v", err)
		return
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Link); err != nil {
			http.Error(w, "Failed to scan article", http.StatusInternalServerError)
			return
		}
		articles = append(articles, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(articles)
}

// Article Fetching and Parsing Logic
func updateArticlesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting article update process...")

	newArticleCount, err := triggerFeedUpdate()
	if err != nil {
		log.Printf("Error during feed update process: %v", err)
		http.Error(w, "Failed to update feeds", http.StatusInternalServerError)
		return
	}

	log.Printf("Article update process completed. Added %d new articles.", newArticleCount)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":             "success",
		"new_articles_count": newArticleCount,
	})
}

// FeedWithID is a helper struct to pass feed data to goroutines.
type FeedWithID struct {
	ID  int
	URL string
}

func triggerFeedUpdate() (int, error) {
	// Get all feeds from the database.
	rows, err := DB.Query("SELECT id, url FROM feeds")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var feedsToFetch []FeedWithID
	for rows.Next() {
		var feed FeedWithID
		if err := rows.Scan(&feed.ID, &feed.URL); err != nil {
			log.Printf("Could not scan feed row: %v", err)
			continue // Skip this feed on scan error
		}
		feedsToFetch = append(feedsToFetch, feed)
	}

	// Concurrently fetch and parse feeds.
	var wg sync.WaitGroup
	itemsChan := make(chan *gofeed.Item, 100) // Buffered channel to hold parsed articles
	feedIDChan := make(chan int, 100)         // Channel to associate items with their feed ID

	for _, feed := range feedsToFetch {
		wg.Add(1)
		go fetchAndParseFeed(feed, &wg, itemsChan, feedIDChan)
	}

	// Wait for all fetches to complete, then close the channels.
	go func() {
		wg.Wait()
		close(itemsChan)
		close(feedIDChan)
	}()

	// Store new articles in the database.
	p := bluemonday.StripTagsPolicy()
	var newArticleCount int
	// We need to consume both channels in parallel
	for item := range itemsChan {
		feedID := <-feedIDChan
		sanitizedContent := p.Sanitize(item.Description) // Sanitize any html from description
		// Use PostgreSQL's "ON CONFLICT" to handle duplicates.
		res, err := DB.Exec(`
			INSERT INTO articles (title, content, link, published_at, feed_id)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (link) DO NOTHING`,
			item.Title, sanitizedContent, item.Link, item.PublishedParsed, feedID,
		)
		if err != nil {
			log.Printf("Error inserting article '%s': %v", item.Title, err)
			continue
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected > 0 {
			newArticleCount++
		}
	}

	return newArticleCount, nil
}

func fetchAndParseFeed(feed FeedWithID, wg *sync.WaitGroup, itemsChan chan<- *gofeed.Item, feedIDChan chan<- int) {
	defer wg.Done()
	fp := gofeed.NewParser()
	parsedFeed, err := fp.ParseURL(feed.URL)
	if err != nil {
		log.Printf("Failed to parse feed from %s: %v", feed.URL, err)
		return
	}

	for _, item := range parsedFeed.Items {
		itemsChan <- item
		feedIDChan <- feed.ID
	}
}