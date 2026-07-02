package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
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
	usableBalance := fiatBalance * 0.97             // keep 3% buffer for fees/slippage/reserved funds
	put := math.Floor((usableBalance/10)*100) / 100 // round down to 2 decimal places
	fmt.Printf("Calculated order size: %.2f %s\n", put, fiat.Currency)

	currentBuyPrice, err := getBuyPrice(fmt.Sprintf("%s-%s", crypto.Currency, fiat.Currency))
	if err != nil {
		log.Fatalf("error getting buy price: %v", err)
	}
	discounts := []float64{0, 0.01, 0.02, 0.03, 0.04, 0.05, 0.06, 0.07, 0.08, 0.10}
	for i, discount := range discounts {
		// putting in 10 orders at a time, each with the same size, but the buy price will decrease by a percentage with each order
		limitPrice := currentBuyPrice * (1 - discount)
		baseSize := put / limitPrice

		fmt.Printf("Put in %.2f for %.2f\n", put, limitPrice)

		// Creating order with the calculated limit price and base size and a unique client order ID using the current timestamp and the index of the order in the loop
		order, err := createOrder(CreateOrderRequest{
			ClientOrderID: fmt.Sprintf("order_%d_%d", time.Now().UnixNano(), i),
			ProductID:     fmt.Sprintf("%s-%s", crypto.Currency, fiat.Currency),
			Side:          "BUY",
			OrderConfiguration: OrderConfiguration{
				LimitLimitGTC: LimitLimitGTC{
					BaseSize:   fmt.Sprintf("%.8f", baseSize),
					LimitPrice: fmt.Sprintf("%.2f", limitPrice),
					PostOnly:   false,
				},
			},
		})
		if err != nil {
			log.Fatalf("error creating order: %v", err)
		}
		fmt.Printf("Order created: %+v\n", order.SuccessResponse.OrderID)
	}

}
