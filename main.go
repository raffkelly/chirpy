package main

import (
	"fmt"
	"net/http"
)

func main() {
	multiplex := http.NewServeMux()
	multiplex.Handle("/", http.FileServer(http.Dir(".")))
	server := http.Server{
		Addr:    ":8080",
		Handler: multiplex,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err.Error())
	}
}
