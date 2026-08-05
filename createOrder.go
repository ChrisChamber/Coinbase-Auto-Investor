package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type CreateOrderRequest struct {
	ClientOrderID      string             `json:"client_order_id"`
	ProductID          string             `json:"product_id"`
	Side               string             `json:"side"`
	OrderConfiguration OrderConfiguration `json:"order_configuration"`
}

type OrderConfiguration struct {
	LimitLimitGTC LimitLimitGTC `json:"limit_limit_gtc"`
}

type LimitLimitGTC struct {
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
	PostOnly   bool   `json:"post_only"`
}

type orderResponse struct {
	Success         bool `json:"success"`
	SuccessResponse struct {
		OrderID       string `json:"order_id"`
		ProductID     string `json:"product_id"`
		Side          string `json:"side"`
		ClientOrderID string `json:"client_order_id"`
	} `json:"success_response"`
	ErrorResponse struct {
		Error                string `json:"error"`
		Message              string `json:"message"`
		ErrorDetails         string `json:"error_details"`
		PreviewFailureReason string `json:"preview_failure_reason"`
	} `json:"error_response"`
}

func getCoinbaseAccounts() (*Account, *Account, error) {
	accounts, fiat, crypto, err := getAccounts("EUR", "BTC")
	if err != nil {
		log.Fatalf("error getting accounts: %v", err)
	}
	if fiat != nil {
		log.Printf("Returned %s account id=%s\n", fiat.Currency, fiat.UUID)
		if fiat.AvailableBalance.Value != "0" {
			log.Printf("FIAT account %s has available balance: %s\n", fiat.Currency, fiat.AvailableBalance.Value)
		} else {
			log.Printf("FIAT account %s has no available balance. Please deposit funds.\n", fiat.Currency)
		}
	} else {
		log.Fatalf("No FIAT account found in accounts: %+v\n", accounts)
	}
	if crypto != nil {
		log.Printf("Returned %s account id=%s\n", crypto.Currency, crypto.UUID)
	} else {
		log.Fatalf("No CRYPTO account found in accounts: %+v\n", accounts)
	}

	return fiat, crypto, nil
}

func calculateOrderSize(fiat *Account) (float64, error) {
	// Dividing available balance by 10 and rounding down to 2 decimal places for the order size
	fiatBalance, err := strconv.ParseFloat(fiat.AvailableBalance.Value, 64)
	if err != nil {
		log.Fatalf("error parsing fiat balance: %v", err)
	}
	usableBalance := fiatBalance * 0.97             // keep 3% buffer for fees/slippage/reserved funds
	put := math.Floor((usableBalance/10)*100) / 100 // round down to 2 decimal places
	log.Printf("Calculated order size: %.2f %s\n", put, fiat.Currency)

	return put, nil
}

func createOrder(order CreateOrderRequest) (orderResponse, error) {

	body, err := json.Marshal(order)
	if err != nil {
		return orderResponse{}, fmt.Errorf("marshal order payload: %w", err)
	}

	req, _ := http.NewRequest("POST", "https://api.coinbase.com/api/v3/brokerage/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+getJwt("POST", "api.coinbase.com", "/api/v3/brokerage/orders"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return orderResponse{}, fmt.Errorf("send order request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return orderResponse{}, fmt.Errorf("create order failed: %s, body: %s", resp.Status, respBody)
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		var orderResp orderResponse
		if err := json.Unmarshal(respBody, &orderResp); err != nil {
			return orderResponse{}, fmt.Errorf("unmarshal order response: %w", err)
		}
		fmt.Printf("Order created successfully: %+v\n", orderResp)
		return orderResp, nil
	}
}

func createBatchOrders(fiat, crypto *Account, amount float64) error {
	currentBuyPrice, err := getBuyPrice(fmt.Sprintf("%s-%s", crypto.Currency, fiat.Currency))
	if err != nil {
		log.Fatalf("error getting buy price: %v", err)
	}
	discounts := []float64{0, 0.005, 0.01, 0.015, 0.02, 0.025, 0.03, 0.035, 0.04, 0.045}
	storedOrders, err := loadStoredOrders()
	if err != nil {
		log.Fatalf("error loading stored orders: %v", err)
	}
	for i, discount := range discounts {
		// putting in 10 orders at a time, each with the same size, but the buy price will decrease by a percentage with each order
		limitPrice := currentBuyPrice * (1 - discount)
		baseSize := amount / limitPrice

		log.Printf("Put in %.2f for %.2f", amount, limitPrice)

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
		log.Printf("Order created: %+v\n", order.SuccessResponse.OrderID)

		// accumulate the order details in the storedOrders slice to save them later
		storedOrders = append(storedOrders, StoredOrder{
			OrderID:       order.SuccessResponse.OrderID,
			ClientOrderID: order.SuccessResponse.ClientOrderID,
			ProductID:     order.SuccessResponse.ProductID,
			Side:          order.SuccessResponse.Side,
			Status:        "OPEN",
		})

	}
	return saveStoredOrders(storedOrders)
}
