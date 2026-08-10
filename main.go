package main

import (
	"log"
	"os"
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
	// Check and retry pending orders before checking balance and creating new orders
	if err := retryPendingOrders(); err != nil {
		log.Fatalf("ERROR: recovering pending orders: %v", err)
	}
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

func runLoop() {
	// Check and retry pending orders on startup
	if err := retryPendingOrders(); err != nil {
		log.Fatalf("ERROR: recovering pending orders: %v", err)
	}
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
				}
			}
		}
	}
}

func main() {
	// append to file instead of truncating, create file if does not exist, open for writing only, give read/write permissions to owner, read permissions to group and others
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
