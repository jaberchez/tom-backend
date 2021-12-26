package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func getEnvsV1(w http.ResponseWriter, r *http.Request) {
	envs := make(map[string]string)

	// Get environment variables
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}

	//w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusCreated)
	//json.NewEncoder(w).Encode(envs)

	jData, err := json.Marshal(envs)

	if err != nil {
		log.Printf("problems marshalling json data: %s", err.Error())

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "some internal error ocurred")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jData)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func main() {
	listenPort := os.Getenv("PORT")

	if len(listenPort) == 0 {
		listenPort = "8080"
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/env", getEnvsV1).Methods("GET") // Only GET allowed
	r.HandleFunc("/startup", healthCheck)
	r.HandleFunc("/liveness", healthCheck)
	r.HandleFunc("/readiness", healthCheck)
	r.HandleFunc("/", healthCheck)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server listening or port %s\n", listenPort)
	log.Fatal(srv.ListenAndServe())
}
