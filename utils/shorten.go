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

    url_encoded := base64.URLEncoding.EncodeToString([]byte(url))

    // Concatenate the byte values of url_encoded and ts_encoded
    combined := append([]byte(url_encoded), []byte(ts_encoded)...)

    // Base64 encode the combined bytes and take the first 8 characters
    final_encoded := base64.URLEncoding.EncodeToString(combined)
    shortened := final_encoded[:8]

    return shortened, nil
}