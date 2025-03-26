package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
)

func handlerReadiness(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	_, err := res.Write([]byte("OK"))
	if err != nil {
		log.Printf("unable to write response body: %v\n", err)
	}

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

/*
func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	a.fileserverHits.Add(1)
	return next
}
*/

func (a *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) handlerMetrics(res http.ResponseWriter, req *http.Request) {
	body := "Hits: " + strconv.Itoa(int(a.fileserverHits.Load()))
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	_, err := res.Write([]byte(body))
	if err != nil {
		log.Printf("unable to write response body: %v\n", err)
	}
}

func (a *apiConfig) handleReset(res http.ResponseWriter, req *http.Request) {
	a.fileserverHits.Store(0)
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(200)
	_, err := res.Write([]byte("Hit counter reset"))
	if err != nil {
		log.Printf("unable to write response body: %v\n", err)
	}
}

func main() {
	apiCfg := &apiConfig{}
	multiplex := http.NewServeMux()
	fileServ := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	multiplex.Handle("/app/", apiCfg.middlewareMetricsInc(fileServ))
	multiplex.HandleFunc("GET /healthz", handlerReadiness)
	multiplex.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	multiplex.HandleFunc("POST /reset", apiCfg.handleReset)
	server := http.Server{
		Addr:    ":8080",
		Handler: multiplex,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}
