package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"time"
)

func ShortenURL(url string) (string, error) {
    if url == "" {
        return "", errors.New("URL cannot be empty")
    }

	db, db_err := sql.Open("sqlite3", "./urls.db")
	if db_err != nil {
		return "", db_err
	}
	defer db.Close()

	// Check if the original_url already exists
	var existing_shortened string
	query := `SELECT shortened_url FROM urls WHERE original_url = ?`
	exists_err := db.QueryRow(query, url).Scan(&existing_shortened)
	if exists_err == nil {
		// Found existing shortcode
		return existing_shortened, nil
	} else if exists_err != sql.ErrNoRows {
		return "", exists_err
	}

	/* shortcode not found, generate a new one
	Generate a unique shortened URL using SHA-256 and Base64 encoding
	Combine the URL with the current timestamp to ensure uniqueness */
    ts := time.Now().UnixNano()
    ts_bytes := []byte(fmt.Sprintf("%d", ts))
    ts_encoded := base64.URLEncoding.EncodeToString(ts_bytes)

    url_encoded := base64.URLEncoding.EncodeToString([]byte(url))

    // Concatenate the byte values of url_encoded and ts_encoded
    combined := append([]byte(url_encoded), []byte(ts_encoded)...)

	// Hash the combined bytes using SHA-256
	hash := sha256.Sum256(combined)

    // Base64 encode the hash and take the first 8 characters
    final_encoded := base64.URLEncoding.EncodeToString(hash[:])
    shortened := final_encoded[:8]

	insert_sql := `INSERT INTO urls (original_url, shortened_url) VALUES (?, ?)`
	_, err := db.Exec(insert_sql, url, shortened)
	if err != nil {
		return "", err
	}
	log.Println("New URL shortened and stored:", url, "->", shortened)

    return shortened, nil
}