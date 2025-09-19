package main

import (
	"database/sql"
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

func main() {
	env_err := godotenv.Load()
	if env_err != nil {
		log.Fatal("Error loading .env file")
	}

	DOMAIN := os.Getenv("DOMAIN")
	if DOMAIN == "" {
		log.Println("DOMAIN environment variable is not set")
		return
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
		// Shorten the provided URL, store it and return it to our UI
		req.ParseForm()
		url := req.FormValue("url")
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "http://" + url
		}
		shortened, err := utils.ShortenURL(url)
		if err != nil {
			http.Error(writer, "Failed to shorten URL: "+err.Error(), http.StatusBadRequest)
			return
		}
		url = DOMAIN + "/" + shortened
		tmpl.ExecuteTemplate(writer, "shorten.html", PageData{
			ShortenedURL: url,
		})
	})

	router.HandleFunc("/", utils.RedirectHandler)

	// Put endpoints above server start
	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server is running on http://localhost:8080")

	serv_err := srv.ListenAndServe()
	if serv_err != nil && !errors.Is(serv_err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", serv_err)
	}

}
