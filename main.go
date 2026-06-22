package main

import (
	"fmt"
	"math"
	"os"
	"strconv"

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
	} else {
		log.Fatalf("No FIAT account found in accounts: %+v\n", accounts)
	}
	if crypto != nil {
		fmt.Printf("Returned %s account id=%s\n", crypto.Currency, crypto.UUID)
	} else {
		log.Fatalf("No CRYPTO account found in accounts: %+v\n", accounts)
	}

	// Dividing available balance by 10 and rounding down to 2 decimal places for the order size
	fiatBalance, err := strconv.ParseFloat(fiat.AvailableBalance.Value, 64)
	if err != nil {
		log.Fatalf("error parsing fiat balance: %v", err)
	}

	put := math.Floor(fiatBalance / 10)
	fmt.Printf("Calculated order size: %.2f %s\n", put, fiat.Currency)

	for range 10 {

	}
}
