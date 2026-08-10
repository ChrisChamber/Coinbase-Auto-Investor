package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
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
		log.Printf("Returned %s account id=%s", fiat.Currency, fiat.UUID)
		if fiat.AvailableBalance.Value != "0" {
			log.Printf("FIAT account %s has available balance: %s", fiat.Currency, fiat.AvailableBalance.Value)
		} else {
			log.Printf("FIAT account %s has no available balance. Please deposit funds.", fiat.Currency)
		}
	} else {
		log.Fatalf("ERROR: No FIAT account found in accounts: %+v", accounts)
	}
	if crypto != nil {
		log.Printf("Returned %s account id=%s", crypto.Currency, crypto.UUID)
	} else {
		log.Fatalf("ERROR: No CRYPTO account found in accounts: %+v", accounts)
	}

	return fiat, crypto, nil
}

func calculateOrderSize(fiat *Account) (decimal.Decimal, error) {
	// Dividing available balance by 10 and rounding down to 2 decimal places for the order size
	fiatBalance, err := decimal.NewFromString(fiat.AvailableBalance.Value)
	if err != nil {
		return decimal.Zero, fmt.Errorf("ERROR: error parsing fiat balance: %v", err)
	}
	usableBalance := fiatBalance.Mul(decimal.NewFromFloat(0.97)) // keep 3% buffer for fees/slippage/reserved funds
	orderSize := usableBalance.Div(decimal.NewFromInt(10))       // divide by 10 to create 10 orders
	log.Printf("Calculated order size: %s %s", orderSize.StringFixed(2), fiat.Currency)
	// round down to 2 decimal places
	return orderSize.Truncate(2), nil
}

func createOrder(order CreateOrderRequest) (orderResponse, error) {

	body, err := json.Marshal(order)
	if err != nil {
		return orderResponse{}, fmt.Errorf("marshal order payload: %w", err)
	}

	req, _ := http.NewRequest("POST", "https://api.coinbase.com/api/v3/brokerage/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+getJwt("POST", "api.coinbase.com", "/api/v3/brokerage/orders"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
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
		log.Printf("Order created successfully: %+v", orderResp)
		return orderResp, nil
	}
}

func createBatchOrders(fiat, crypto *Account, amount decimal.Decimal) error {
	currentBuyPrice, err := getBuyPrice(fmt.Sprintf("%s-%s", crypto.Currency, fiat.Currency))
	if err != nil {
		log.Fatalf("ERROR: error getting buy price: %v", err)
	}
	discounts := []decimal.Decimal{
		decimal.NewFromFloat(0),
		decimal.NewFromFloat(0.005),
		decimal.NewFromFloat(0.01),
		decimal.NewFromFloat(0.015),
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(0.025),
		decimal.NewFromFloat(0.03),
		decimal.NewFromFloat(0.035),
		decimal.NewFromFloat(0.04),
		decimal.NewFromFloat(0.045)}

	storedOrders, err := loadStoredOrders()
	if err != nil {
		log.Fatalf("ERROR: error loading stored orders: %v", err)
	}
	for i, discount := range discounts {
		// putting in 10 orders at a time, each with the same size, but the buy price will decrease by a percentage with each order
		limitPrice := currentBuyPrice.Mul(decimal.NewFromInt(1).Sub(discount))
		baseSize := amount.Div(limitPrice)
		clientOrderID := fmt.Sprintf("order_%d_%d", time.Now().UnixNano(), i)

		// Building order request before calling coinbase API to create the order
		log.Printf("Put in %.2f for %.2f", amount, limitPrice)
		storedOrder := StoredOrder{
			ClientOrderID: clientOrderID,
			ProductID:     fmt.Sprintf("%s-%s", crypto.Currency, fiat.Currency),
			Side:          "BUY",
			Status:        "PENDING_SUBMISSION",
			BaseSize:      baseSize.StringFixed(8),
			LimitPrice:    limitPrice.StringFixed(2),
		}
		storedOrders = append(storedOrders, storedOrder)

		// Save the order immediately after creating it to ensure that we have a record of it in case of any failures
		if err := saveStoredOrders(storedOrders); err != nil {
			return fmt.Errorf("save pending orders: %w", err)
		}

		// Creating order with the calculated limit price and base size and a unique client order ID using the current timestamp and the index of the order in the loop
		order, err := createOrder(CreateOrderRequest{
			ClientOrderID: clientOrderID,
			ProductID:     storedOrder.ProductID,
			Side:          "BUY",
			OrderConfiguration: OrderConfiguration{
				LimitLimitGTC: LimitLimitGTC{
					BaseSize:   storedOrder.BaseSize,
					LimitPrice: storedOrder.LimitPrice,
					PostOnly:   false,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		storedOrders[len(storedOrders)-1].OrderID = order.SuccessResponse.OrderID
		storedOrders[len(storedOrders)-1].Status = "OPEN"
		// Once order is successfully opened, we can save the order details to the storedOrders slice and persist it to the JSON file
		if err := saveStoredOrders(storedOrders); err != nil {
			return fmt.Errorf("save confirmed order: %w", err)
		}

	}
	if err := sendNotification("BTC Orders Created", fmt.Sprintf("%d orders created successfully.", len(discounts))); err != nil {
		log.Printf("ERROR: sending notification: %v", err)
	}
	return nil
}
