package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var (
	isServerReady bool
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

func startupHealthCheck(w http.ResponseWriter, r *http.Request) {
	if isServerReady {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Listener is up and running")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Listener is not ready")
	}
}

func main() {
	listenPort := os.Getenv("PORT")

	if len(listenPort) == 0 {
		listenPort = "8080"
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/env", getEnvsV1).Methods("GET") // Only GET allowed
	r.HandleFunc("/startup", startupHealthCheck)
	r.HandleFunc("/liveness", healthCheck)
	r.HandleFunc("/readiness", healthCheck)
	r.HandleFunc("/", healthCheck)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", listenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// Unexpected error, port in use?
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	log.Printf("Server listening or port %s\n", listenPort)

	isServerReady = true

	stopC := make(chan os.Signal)
	signal.Notify(stopC, syscall.SIGTERM, syscall.SIGINT)
	sig := <-stopC

	// For health checks
	isServerReady = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch sig {
	case syscall.SIGTERM:
		log.Println("got signal SIGTERM")
	case syscall.SIGINT:
		log.Println("got signal SIGINT")
	default:
		log.Println("got unknown signal")
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err) // Failure/timeout shutting down the server gracefully
	}

	log.Println("server exited properly")
}
