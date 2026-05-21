//go:build ignore

package main

import (
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	log.Println("Local static web server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
