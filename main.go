package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()
	sm, err := NewSandboxManager()
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if err != nil {
		log.Fatal(err)
	}

	id, err := sm.CreateSandbox(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sandbox created: ", id)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}