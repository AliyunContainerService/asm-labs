package main

import (
	"fmt"
	"net/http"

	"github.com/asm-labs/llm-token-ratelimit-service-example/pkg/handlers"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ratelimit", handlers.RateLimit)
	r.HandleFunc("/update_ratelimit_record", handlers.UpdateRateLimitRecord)

	http.Handle("/", r)
	fmt.Println("Starting server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
