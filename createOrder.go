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
