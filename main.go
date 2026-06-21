package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("missing required environment variable: %s", name)
	}
	return value
}

func getKeyName() string {
	return mustGetEnv("COINBASE_KEY_NAME")
}

func getKeySecret() string {
	path := mustGetEnv("COINBASE_KEY_SECRET")

	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read private key file %s: %v", path, err)
	}

	return string(b)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	accounts, eur, btc, err := getAccounts()
	if err != nil {
		log.Fatalf("error getting accounts: %v", err)
	}
	fmt.Printf("Retrieved %d accounts on final page\n", accounts.Size)
	if eur != nil {
		fmt.Printf("Returned EUR account id=%s\n", eur.UUID)
	}
	if btc != nil {
		fmt.Printf("Returned BTC account id=%s\n", btc.UUID)
	}

}
