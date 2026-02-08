package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	log.Printf("Starting product microservice")
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQUEST: %v", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Hello world!"}); err != nil {
			log.Printf("ERROR: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	log.Fatal(http.ListenAndServe(":5000", nil))
}
