package main

import (
	"os"
	"time"

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

func getPushoverToken() string {
	return mustGetEnv("PUSHOVER_TOKEN")
}

func getPushoverUser() string {
	return mustGetEnv("PUSHOVER_USER")
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

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
}

func checkBalanceandCreateOrders() error {
	canCreateOrders, err := canCreatenewBatchOrders()
	if err != nil {
		log.Errorf("error checking if orders can be created: %v", err)
	}

	if !canCreateOrders {
		log.Println("Orders cannot be created at this time. Please check your account balance and existing orders.")
		return nil
	}

	fiat, crypto, err := getCoinbaseAccounts()
	if err != nil {
		log.Fatalf("error getting coinbase accounts: %v", err)
	}

	amount, err := calculateOrderSize(fiat)
	if err != nil {
		log.Fatalf("error calculating order size: %v", err)
	}

	if amount <= 0 {
		log.Println("Insufficient funds to create orders.")
		return nil

	}

	if err := createBatchOrders(fiat, crypto, amount); err != nil {
		log.Fatalf("error creating batch orders: %v", err)
	}

	return nil
}

func runLoop() {
	if err := checkBalanceandCreateOrders(); err != nil {
		log.Fatalf("error checking balance and creating orders: %v", err)
	}
	moneyTicker := time.NewTicker(24 * time.Hour)
	orderTicker := time.NewTicker(1 * time.Hour)
	defer moneyTicker.Stop()
	defer orderTicker.Stop()

	for {
		select {
		case <-moneyTicker.C:
			if err := checkBalanceandCreateOrders(); err != nil {
				log.Fatalf("error checking balance and creating orders: %v", err)
			}
		case <-orderTicker.C:
			ordersFilled, err := checkIfAllOrdersFilled()
			if err != nil {
				log.Fatalf("error checking orders: %v", err)
			}

			if ordersFilled {
				log.Println("All orders are filled. Proceed with sending to wallet.")
				if err := sendNotification("Orders Filled", "All orders are filled successfully. Proceed with transfer."); err != nil {
					log.Errorf("error sending notification: %v", err)
				}
			}
		}
	}
}

func main() {
	loadEnv()
	runLoop()
}
