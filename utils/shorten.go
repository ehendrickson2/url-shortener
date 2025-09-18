package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

func ShortenURL(url string) (string, error) {
	if url == "" {
		return "", errors.New("URL cannot be empty")
	}
	// Use current timestamp to ensure uniqueness
	ts := time.Now().UnixNano()
	ts_bytes := []byte(fmt.Sprintf("%d", ts))
	ts_encoded := base64.URLEncoding.EncodeToString(ts_bytes)

	// Simple shortening logic: base64 encode the URL and take the first 8 characters
	encoded := base64.URLEncoding.EncodeToString([]byte(url)) + ts_encoded
	encoded = encoded[:len(encoded)-2] // Remove padding
	shortened := encoded[:8]

	return shortened, nil
}