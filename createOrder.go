package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func createOrder(order CreateOrderRequest) error {

	body, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order payload: %w", err)
	}

	req, _ := http.NewRequest("POST", "https://api.coinbase.com/api/v3/brokerage/orders", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+getJwt("POST", "api.coinbase.com", "/api/v3/brokerage/orders"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return fmt.Errorf("send order request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create order failed: %s", resp.Status)
	} else {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Println(string(respBody))
	}
	return nil
}
