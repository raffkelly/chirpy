package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}
	multiplex := http.NewServeMux()
	fileServ := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	multiplex.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))
	multiplex.HandleFunc("GET /api/healthz", handlerReadiness)
	multiplex.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	multiplex.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	multiplex.HandleFunc("POST /api/validate_chirp", handlerValidate_Chirp)
	server := http.Server{
		Addr:    ":8080",
		Handler: multiplex,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}
