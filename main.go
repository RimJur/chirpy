package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("POST /api/validate_chirp", cfg.handleValidateChirp)
	mux.HandleFunc("GET /api/healthz", handleHealtz)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	server.ListenAndServe()
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handleHealtz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	count := strconv.Itoa(int(cfg.fileserverHits.Load()))
	tmpl, err := template.ParseFiles("./admin-metrics.html")
	if err != nil {
		log.Fatalf("error parsing template: %v", err)
	}

	err = tmpl.Execute(w, count)
	if err != nil {
		log.Fatalf("error executing template: %v", err)
	}
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("Hits: %v", strconv.Itoa(int(cfg.fileserverHits.Swap(0))))
	w.Write([]byte(resp))
}

func (cfg *apiConfig) handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type requestVals struct {
		Body string `json:"body"`
	}
	type responseVals struct {
		Valid bool `json:"valid"`
	}
	reqBody := requestVals{}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&reqBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	if len(reqBody.Body) > 400 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	respBody := responseVals{Valid: true}
	respondWithJSON(w, http.StatusOK, respBody)
}
