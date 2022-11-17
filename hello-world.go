package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	// handle route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>Hello World!</h1>")
	})

	// print a log message
	log.Println("Starting server on port 8080")

	// start the server
	http.ListenAndServe(":8080", nil)
}
