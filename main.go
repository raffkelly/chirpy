package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/raffkelly/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func main() {

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to connect to open database connection")
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{}
	apiCfg.dbQueries = dbQueries
	apiCfg.platform = platform

	multiplex := http.NewServeMux()
	fileServ := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	multiplex.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))
	multiplex.HandleFunc("GET /api/healthz", handlerReadiness)
	multiplex.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	multiplex.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	multiplex.HandleFunc("POST /api/validate_chirp", handlerValidate_Chirp)
	multiplex.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	server := http.Server{
		Addr:    ":8080",
		Handler: multiplex,
	}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}
