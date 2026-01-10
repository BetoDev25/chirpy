package main

import (
	"net/http"
	"log"
	"sync/atomic"
	"fmt"
	"encoding/json"
	"database/sql"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/BetoDev25/chirpy/internal/database"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorMsg struct {
		Error string `json:"error"`
	}

	resp := errorMsg{
		Error: msg,
	}
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(500)
		return
	} else {
		w.WriteHeader(code)
	}

	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(payload)
}

type apiConfig struct {
	fileserverHits atomic.Int32
	db       *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Platform is not dev")
		return
	}
	err := cfg.db.Reset(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't reset")
		return
	}
	cfg.fileserverHits.Store(0)

	respondWithJSON(w, http.StatusOK, struct{}{})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	body := fmt.Sprintf(`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, hits)
	w.Write([]byte(body))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	plat := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	serve := http.NewServeMux()
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       plat,
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: serve,
	}

	serve.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	serve.HandleFunc("GET /api/healthz", handlerHealthz)
	serve.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serve.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	serve.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	serve.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	serve.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	serve.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	serve.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
