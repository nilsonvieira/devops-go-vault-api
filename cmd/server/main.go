package main

import (
	"devops-go-vault-api/config"
	"devops-go-vault-api/internal/handler"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	config.LoadConfig()

	router := mux.NewRouter()
	router.HandleFunc("/sendVault", handler.StoreHandler).Methods("POST")
	router.HandleFunc("/convert", handler.ConvertHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
