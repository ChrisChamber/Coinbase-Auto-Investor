package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("ERROR: missing required environment variable: %s", name)
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
		log.Fatalf("ERROR: failed to read private key file %s: %v", path, err)
	}

	return string(b)
}

func checkBalanceandCreateOrders() error {
	canCreateOrders, err := canCreatenewBatchOrders()
	if err != nil {
		log.Printf("ERROR: checking if orders can be created: %v", err)
	}

	if !canCreateOrders {
		log.Println("ERROR: Orders cannot be created at this time. Please check your account balance and existing orders.")
		return nil
	}

	fiat, crypto, err := getCoinbaseAccounts()
	if err != nil {
		log.Fatalf("ERROR: getting coinbase accounts: %v", err)
	}

	amount, err := calculateOrderSize(fiat)
	if err != nil {
		log.Fatalf("ERROR: calculating order size: %v", err)
	}

	if amount <= 0 {
		log.Println("ERROR: Insufficient funds to create orders.")
		return nil

	}

	if err := createBatchOrders(fiat, crypto, amount); err != nil {
		log.Fatalf("ERROR: creating batch orders: %v", err)
	}

	return nil
}

func createLogFile(logDirectory string) (*os.File, error) {
	if err := os.MkdirAll(logDirectory, 0750); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	logPath := filepath.Join(
		logDirectory,
		"bot-"+time.Now().Format("2006-01-02")+".log",
	)

	logFile, err := os.OpenFile(
		logPath,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0640,
	)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	return logFile, nil
}

func runLoop() {
	if err := checkBalanceandCreateOrders(); err != nil {
		log.Fatalf("ERROR: checking balance and creating orders: %v", err)
	}
	moneyTicker := time.NewTicker(24 * time.Hour)
	orderTicker := time.NewTicker(1 * time.Hour)
	defer moneyTicker.Stop()
	defer orderTicker.Stop()

	for {
		select {
		case <-moneyTicker.C:
			if err := checkBalanceandCreateOrders(); err != nil {
				log.Fatalf("ERROR: checking balance and creating orders: %v", err)
			}
		case <-orderTicker.C:
			ordersFilled, err := checkIfAllOrdersFilled()
			if err != nil {
				log.Fatalf("ERROR: checking orders: %v", err)
			}

			if ordersFilled {
				log.Println("All orders are filled. Proceed with sending to wallet.")
				if err := sendNotification("Orders Filled", "All orders are filled successfully. Proceed with transfer."); err != nil {
					log.Printf("ERROR: sending notification: %v", err)
				}
			}
		}
	}
}

func main() {
	runLoop()
}
