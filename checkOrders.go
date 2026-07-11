package main

import (
	"encoding/json"
	"os"
)

type StoredOrder struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Status        string `json:"status"`
}

const ordersFile = "orders.json"

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
