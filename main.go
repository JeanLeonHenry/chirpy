package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		log.Printf("Incremented fileServerHits to %v", cfg.fileserverHits)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprint("Hits: ", cfg.fileserverHits)))
}
func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}
func main() {
	mux := http.NewServeMux()
	server := http.Server{Addr: ":8080", Handler: mux}
	cfg := apiConfig{fileserverHits: 0}
	rootServer := http.FileServer(http.Dir("."))

	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", rootServer)))
	mux.HandleFunc("/healthz", handlerReadiness)
	mux.HandleFunc("/metrics", cfg.handlerMetrics)
	mux.HandleFunc("/reset", cfg.handlerResetMetrics)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
