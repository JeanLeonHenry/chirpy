package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JeanLeonHenry/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	data, _ := json.Marshal(payload)
	w.Write(data)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func NewConfig(path string) apiConfig {
	db, _ := database.NewDB(path)
	return apiConfig{db: db}
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{Addr: ":8080", Handler: mux}
	cfg := NewConfig("database.json")
	rootServer := http.FileServer(http.Dir("."))

	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", rootServer)))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("/api/reset", cfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/chirps", cfg.handlerSaveChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerReadChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerReadChirpById)
	log.Fatal(server.ListenAndServe())
}
