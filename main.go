package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/ehendrickson2/url-shortener/utils"
	"github.com/joho/godotenv"
)

/* PageData holds data to be passed to templates
Look into putting structs into project_root/models/ */
type PageData struct {
	Name string
}

func main() {
	env_err := godotenv.Load()
	if env_err != nil {
		log.Fatal("Error loading .env file")
	}
	
	// Load templates
	tmpl := template.Must(template.New("").ParseGlob("templates/*.html"))

	router := http.NewServeMux()

	router.HandleFunc("GET /{$}", func(writer http.ResponseWriter, req *http.Request) {
		tmpl.ExecuteTemplate(writer, "index.html", PageData{
			Name: "User!",
		})
	})

	router.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		// Shorten the provided URL, store it and return it to our UI
		DOMAIN := os.Getenv("DOMAIN")
		req.ParseForm()
		url := req.FormValue("url")
		shortened, err := utils.ShortenURL(url)
		if err != nil {
			http.Error(writer, "Failed to shorten URL: "+err.Error(), http.StatusBadRequest)
			return
		}
		url = DOMAIN + "/" + shortened
		fmt.Fprintf(writer, "Shortened URL: %s", url)
	})

	// Put endpoints above server start
	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server is running on http://localhost:8080")

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Error starting server:", err)
	}
	
}