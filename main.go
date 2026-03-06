package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ehendrickson2/url-shortener/utils"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

/*
	PageData holds data to be passed to templates

Look into putting structs into project_root/models/
*/
type PageData struct {
	ShortenedURL string
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func normalizeConfiguredBaseURL(raw string) string {
	base := strings.TrimSpace(raw)
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		return ""
	}

	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}

	return base
}

func normalizeURL(rawURL string) string {
	url := strings.TrimSpace(rawURL)
	if url == "" {
		return ""
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	return url
}

func buildPublicBaseURL(req *http.Request, configuredBaseURL string) string {
	if configuredBaseURL != "" {
		return configuredBaseURL
	}

	scheme := "http"
	if req != nil {
		if forwardedProto := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
			scheme = strings.Split(forwardedProto, ",")[0]
		} else if req.TLS != nil {
			scheme = "https"
		}

		if req.Host != "" {
			return fmt.Sprintf("%s://%s", scheme, req.Host)
		}
	}

	return "http://localhost:8080"
}

func parseAllowedOrigins(originsEnv string) map[string]bool {
	allowed := map[string]bool{}
	for _, origin := range strings.Split(originsEnv, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowed[trimmed] = true
		}
	}

	return allowed
}

func setCORSHeaders(writer http.ResponseWriter, req *http.Request, allowedOrigins map[string]bool) {
	origin := strings.TrimSpace(req.Header.Get("Origin"))
	if origin == "" {
		return
	}

	if allowedOrigins[origin] {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Vary", "Origin")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	baseURL := normalizeConfiguredBaseURL(os.Getenv("BASE_URL"))
	baseURLSource := "BASE_URL"
	if baseURL == "" {
		baseURL = normalizeConfiguredBaseURL(os.Getenv("DOMAIN"))
		baseURLSource = "DOMAIN"
	}
	if baseURL == "" {
		baseURL = normalizeConfiguredBaseURL(os.Getenv("RENDER_EXTERNAL_URL"))
		baseURLSource = "RENDER_EXTERNAL_URL"
	}

	allowedOrigins := parseAllowedOrigins(os.Getenv("CORS_ORIGINS"))
	if len(allowedOrigins) == 0 {
		allowedOrigins = parseAllowedOrigins("https://eddiehendrickson.com,https://www.eddiehendrickson.com,http://localhost:8000,http://localhost:3000")
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		log.Println("PORT environment variable is not set, defaulting to 8080")
		PORT = "8080"
	}

	db, db_err := sql.Open("sqlite3", "./urls.db")
	if db_err != nil {
		log.Fatal("Error opening database:", db_err)
	}
	defer db.Close()

	create_table_sql := `CREATE TABLE IF NOT EXISTS urls (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"original_url" TEXT UNIQUE,
		"shortened_url" TEXT
	  );`

	_, table_err := db.Exec(create_table_sql)
	if table_err != nil {
		log.Fatal("Error creating table:", table_err)
	}
	log.Println("Database and table initialized.")

	// Load templates
	tmpl := template.Must(template.New("").ParseGlob("templates/*.html"))

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(writer http.ResponseWriter, req *http.Request) {
		tmpl.ExecuteTemplate(writer, "index.html", nil)
	})

	router.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Shorten the provided URL, store it and return it to our UI
		req.ParseForm()
		url := normalizeURL(req.FormValue("url"))
		if url == "" {
			http.Error(writer, "URL cannot be empty", http.StatusBadRequest)
			return
		}
		shortened, err := utils.ShortenURL(url)
		if err != nil {
			http.Error(writer, "Failed to shorten URL: "+err.Error(), http.StatusBadRequest)
			return
		}
		publicBaseURL := buildPublicBaseURL(req, baseURL)
		url = publicBaseURL + "/" + shortened
		tmpl.ExecuteTemplate(writer, "shorten.html", PageData{
			ShortenedURL: url,
		})
	})

	router.HandleFunc("/api/shorten", func(writer http.ResponseWriter, req *http.Request) {
		setCORSHeaders(writer, req, allowedOrigins)

		if req.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		if req.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		var payload shortenRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(writer, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		url := normalizeURL(payload.URL)
		if url == "" {
			http.Error(writer, "URL cannot be empty", http.StatusBadRequest)
			return
		}

		shortened, err := utils.ShortenURL(url)
		if err != nil {
			http.Error(writer, "Failed to shorten URL: "+err.Error(), http.StatusBadRequest)
			return
		}

		publicBaseURL := buildPublicBaseURL(req, baseURL)
		response := shortenResponse{ShortURL: publicBaseURL + "/" + shortened}
		if err := json.NewEncoder(writer).Encode(response); err != nil {
			http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	router.HandleFunc("/", utils.RedirectHandler)

	// Put endpoints above server start
	srv := http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}

	if baseURL != "" {
		log.Printf("Public base URL configured from %s: %s", baseURLSource, baseURL)
	} else {
		log.Println("No BASE_URL/DOMAIN/RENDER_EXTERNAL_URL configured, deriving public URL from request host")
	}

	log.Println("Server is running at", buildPublicBaseURL(nil, baseURL))

	serv_err := srv.ListenAndServe()
	if serv_err != nil && !errors.Is(serv_err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", serv_err)
	}

}
