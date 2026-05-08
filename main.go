package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/etuhoha/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (api *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		api.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (api *apiConfig) metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		template := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
		`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		html := fmt.Sprintf(template, api.fileserverHits.Load())
		w.Write([]byte(html))
	})
}

func (api *apiConfig) resetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		api.fileserverHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func main() {
	apiConfig := apiConfig{}

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("can't load the environment: %s\n", err)
		os.Exit(1)
	}

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("can't open DB: %s\n", err)
		os.Exit(1)
	}
	apiConfig.db = database.New(db)

	mux := http.NewServeMux()

	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	// front
	mux.Handle("/app/", apiConfig.middlewareMetricsInc(fileHandler))

	// public API
	mux.HandleFunc("GET /api/healthz", handleHelthz)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	// admin API
	mux.Handle("GET /admin/metrics", apiConfig.metricsHandler())
	mux.Handle("POST /admin/reset", apiConfig.resetHandler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Printf("Serving at %s...\n", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}

func handleHelthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
