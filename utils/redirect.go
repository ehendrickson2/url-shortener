package utils

import (
	"database/sql"
	"net/http"
)

func RedirectHandler(writer http.ResponseWriter, req *http.Request) {
	shortened_url := req.URL.Path[1:]
	if shortened_url == "" || shortened_url == "shorten" {
		http.Error(writer, "Shortened URL not provided", http.StatusBadRequest)
		return
	}

	db, db_err := sql.Open("sqlite3", "./urls.db")
	if db_err != nil {
		http.Error(writer, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var original_url string
	query := `SELECT original_url FROM urls WHERE shortened_url = ?`
	err := db.QueryRow(query, shortened_url).Scan(&original_url)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(writer, "Shortened URL not found", http.StatusNotFound)
		} else {
			http.Error(writer, "Database query error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(writer, req, original_url, http.StatusFound)
}