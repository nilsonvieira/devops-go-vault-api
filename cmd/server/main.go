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
	router.HandleFunc("/decSecret", handler.DecryptSecretHandler).Methods("POST")
	router.HandleFunc("/generate", handler.GenerateHandler).Methods("POST")
	router.HandleFunc("/deleteSecret", handler.DeleteSecretHandler).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}
