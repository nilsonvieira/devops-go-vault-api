package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var VaultToken string
var VaultAddress string

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao Carregar Arquivo .env: %v", err)
	}

	VaultToken = os.Getenv("VAULT_TOKEN")
	if VaultToken == "" {
		log.Fatalf("VAULT_TOKEN não foi definido no arquivo .env")
	}

	VaultAddress = os.Getenv("VAULT_ADDRESS")
	if VaultAddress == "" {
		log.Fatalf("VAULT_ADDRESS não foi definido no arquivo .env")
	}
}
