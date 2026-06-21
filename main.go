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

	accounts, fiat, crypto, err := getAccounts("EUR", "BTC")
	if err != nil {
		log.Fatalf("error getting accounts: %v", err)
	}
	if fiat != nil {
		fmt.Printf("Returned %s account id=%s\n", fiat.Currency, fiat.UUID)
		if fiat.AvailableBalance.Value != "0" {
			fmt.Printf("FIAT account %s has available balance: %s\n", fiat.Currency, fiat.AvailableBalance.Value)
		} else {
			log.Fatalf("FIAT account %s has no available balance. Please deposit funds.\n", fiat.Currency)
		}
	}
	if crypto != nil {
		fmt.Printf("Returned %s account id=%s\n", crypto.Currency, crypto.UUID)
	}

}
