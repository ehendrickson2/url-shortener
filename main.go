package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/ehendrickson2/url-shortener/utils"
)

/* PageData holds data to be passed to templates
Look into putting structs into root/models/ */
type PageData struct {
	Name string
}

func main() {
	tmpl := template.Must(template.New("").ParseGlob("templates/*.html"))

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(writer http.ResponseWriter, req *http.Request) {
		tmpl.ExecuteTemplate(writer, "index.html", PageData{
			Name: "User!",
		})
	})

	router.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		// Shorten the provided URL, store it and return it to our UI
		req.ParseForm()
		url := req.FormValue("url")
		shortened, err := utils.ShortenURL(url)
		if err != nil {
			http.Error(writer, "Failed to shorten URL: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(writer, "Shortened URL: %s", shortened)
	})

	// Put endpoints above server start
	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server is running on http://localhost:8080")

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Error starting server:", err)
	}
	
}