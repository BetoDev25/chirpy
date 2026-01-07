package main

import (
	"net/http"
	"log"
	"sync/atomic"
	"fmt"
	"encoding/json"
	"strings"
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
	database       *database.Queries
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
	err := cfg.database.Reset(r.Context())
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

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	} else if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	words := strings.Split(params.Body, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			words[i] = "****"
		}
	}
	cleanSentence := strings.Join(words, " ")
	respondWithJSON(w, 200, map[string]string{"cleaned_body": cleanSentence})
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
		database:       dbQueries,
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
	serve.HandleFunc("POST /api/validate_chirp", apiCfg.handlerChirp)
	serve.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
