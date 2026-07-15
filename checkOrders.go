package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type StoredOrder struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Status        string `json:"status"`
}

type getOrderResponse struct {
	Order struct {
		Status string `json:"status"`
	} `json:"order"`
}

const ordersFile = "orders.json"

// loadStoredOrders reads the stored orders from the local JSON file and returns them as a slice of StoredOrder structs.
func loadStoredOrders() ([]StoredOrder, error) {
	b, err := os.ReadFile(ordersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []StoredOrder{}, nil
		}
		return nil, err
	}

	var orders []StoredOrder
	if err := json.Unmarshal(b, &orders); err != nil {
		return nil, err
	}

	return orders, nil
}

func saveStoredOrders(orders []StoredOrder) error {
	b, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ordersFile, b, 0600)
}

// getOrderStatus retrieves the status of an order by its ID from the Coinbase API.
func getOrderStatus(orderID string) (string, error) {
	path := "/api/v3/brokerage/orders/historical/" + orderID

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.coinbase.com"+path,
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization",
		"Bearer "+getJwt(http.MethodGet, "api.coinbase.com", path),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get order failed: %s: %s", resp.Status, body)
	}

	var result getOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Order.Status, nil
}

// refreshStoredOrders checks the status of all stored orders and updates their status in the local storage.
func refreshStoredOrders() error {
	orders, err := loadStoredOrders()
	if err != nil {
		return err
	}

	for i := range orders {
		status, err := getOrderStatus(orders[i].OrderID)
		if err != nil {
			return fmt.Errorf("checking %s: %w", orders[i].OrderID, err)
		}

		orders[i].Status = status
		fmt.Printf("%s: %s\n", orders[i].OrderID, status)
	}

	return saveStoredOrders(orders)
}
