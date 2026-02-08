package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chuuch/product-microservice/config"
)

func main() {
	log.Printf("Starting product microservice")

	c, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}

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
	log.Printf("Server is listening on port: %v", c.Server.Port)
	log.Fatal(http.ListenAndServe(c.Server.Port, nil))
}
