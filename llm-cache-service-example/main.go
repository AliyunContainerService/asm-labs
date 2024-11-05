package main

import (
	"fmt"
	"net/http"

	"github.com/asm-labs/llm-cache-service-example/pkg/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/lookup", handlers.Lookup)
	r.HandleFunc("/update", handlers.Update)

	http.Handle("/", r)
	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
