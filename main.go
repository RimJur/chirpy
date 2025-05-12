package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	mux.Handle("/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("/healthz", handleHealtz)
	mux.HandleFunc("/metrics", cfg.handlerMetrics)
	mux.HandleFunc("/reset", cfg.handlerReset)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}
	server.ListenAndServe()
}

func handleHealtz(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	resp := fmt.Sprintf("Hits: %v", strconv.Itoa(int(cfg.fileserverHits.Load())))
	w.Write([]byte(resp))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	resp := fmt.Sprintf("Hits: %v", strconv.Itoa(int(cfg.fileserverHits.Swap(0))))
	w.Write([]byte(resp))
}
