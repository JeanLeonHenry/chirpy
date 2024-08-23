package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type chirp struct {
	Body string `json:"body"`
}

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
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	template := `<html>
<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`
	w.Write([]byte(fmt.Sprintf(template, cfg.fileserverHits)))
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	log.Printf("Reset fileServerHits to %v", cfg.fileserverHits)
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func handlerValidation(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print("Error reading request body: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := new(chirp)
	err = json.Unmarshal(data, c)
	if err != nil {
		log.Print("Error unmarshaling request body: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Print("Validating: ", c.Body, len(c.Body))
	if len(c.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"Chirp is too long. Max length is 140."}`))
		return
	}
	w.Write([]byte(`{"valid":true}`))
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{Addr: ":8080", Handler: mux}
	cfg := apiConfig{fileserverHits: 0}
	rootServer := http.FileServer(http.Dir("."))

	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", rootServer)))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("/api/reset", cfg.handlerResetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidation)
	log.Fatal(server.ListenAndServe())
}
