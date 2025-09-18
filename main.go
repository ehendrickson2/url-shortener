package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
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

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("Server is running on http://localhost:8080")

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Error starting server:", err)
	}
	

	http.HandleFunc("/shorten", func(writer http.ResponseWriter, req *http.Request) {
		// Shorten the provided URL, store it and return it to our UI
	})
}