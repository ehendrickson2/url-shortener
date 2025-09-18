package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(writer, "Welcome to the URL Shortener!")
	})

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

	http.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		// Shorten the provided URL, store it and return it to our UI
	})
}