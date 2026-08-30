package main

import (
	"log"
	"os"
	"time"

	"github.com/shopspring/decimal"
)

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("ERROR: missing required environment variable: %s", name)
	}
	return value
}

func checkBalanceandCreateOrders() error {
	// Check and retry pending orders before checking balance and creating new orders
	if err := retryPendingOrders(); err != nil {
		log.Fatalf("ERROR: recovering pending orders: %v", err)
	}

	canCreateOrders, err := canCreatenewBatchOrders()
	if err != nil {
		log.Printf("ERROR: checking if orders can be created: %v", err)
	}

	if !canCreateOrders {
		log.Println("There are insufficient funds or existing orders to create new ones.")
		return nil
	}

	fiat, crypto, err := getCoinbaseAccounts()
	if err != nil {
		log.Fatalf("ERROR: getting coinbase accounts: %v", err)
	}

	fiatBalance, err := decimal.NewFromString(fiat.AvailableBalance.Value)
	if err != nil {
		log.Printf("ERROR: invalid fiat balance: %v", err)
		return nil
	}
	// Checking if fiat balance meets the minimum required amount to create orders
	minFiatBalance, err := decimal.NewFromString(mustGetEnv("MIN_FIAT_BALANCE"))
	if err != nil {
		log.Printf("ERROR: invalid minimum fiat balance: %v", err)
		return nil
	}
	if fiatBalance.Cmp(minFiatBalance) < 0 {
		log.Printf("Available fiat balance is below the minimum required amount of %s. Current balance: %s", mustGetEnv("MIN_FIAT_BALANCE"), fiatBalance)
		return nil
	}

	amount, err := calculateOrderSize(fiat)
	if err != nil {
		log.Fatalf("ERROR: calculating order size: %v", err)
	}

	if err := createBatchOrders(fiat, crypto, amount); err != nil {
		log.Fatalf("ERROR: creating batch orders: %v", err)
	}

	return nil
}

func runLoop() {
	// Check balance and create orders immediately on startup
	if err := checkBalanceandCreateOrders(); err != nil {
		log.Fatalf("ERROR: checking balance and creating orders: %v", err)
	}
	// Check balance and create orders every 24 hours
	moneyTicker := time.NewTicker(24 * time.Hour)
	// Check orders every hour and send notification if all orders are filled
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
					continue
				}
				if err := markAllOrdersCompletionNotified(); err != nil {
					log.Printf("ERROR: saving completion notification state: %v", err)
				}
			}
		}
	}
}

func main() {
	// append to logfile instead of truncating, create logfile if does not exist, open for writing only, give read/write permissions to owner, read permissions to group and others
	f, err := os.OpenFile("CBAutoInvestor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// ensure the file is closed when the program exits
	defer f.Close()
	log.SetOutput(f)
	// set log flags to include date, time, and file line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runLoop()
}
