package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type StoredOrder struct {
	OrderID       string `json:"order_id"`
	ClientOrderID string `json:"client_order_id"`
	ProductID     string `json:"product_id"`
	Side          string `json:"side"`
	Status        string `json:"status"`

	//add field with base size and limit price to be able to recreate the order if it fails
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
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
			log.Printf("ERROR: checking %s: %v", orders[i].OrderID, err)
			return fmt.Errorf("checking %s: %w", orders[i].OrderID, err)
		}

		// Log a message when an order transitions from a non-filled status to "FILLED"
		previousStatus := orders[i].Status
		orders[i].Status = status

		if previousStatus != "FILLED" && status == "FILLED" {
			log.Printf(
				"Order filled: id=%s product=%s side=%s",
				orders[i].OrderID,
				orders[i].ProductID,
				orders[i].Side,
			)
		}
	}

	return saveStoredOrders(orders)
}

// TODO: Add a function to check if all orders are filled and return a boolean value.
func checkIfAllOrdersFilled() (bool, error) {
	if err := refreshStoredOrders(); err != nil {
		return false, err
	}
	orders, err := loadStoredOrders()
	if err != nil {
		return false, err
	}

	// No orders does not mean all orders are filled, it means there are no orders to check
	if len(orders) == 0 {
		return false, nil
	}
	for _, order := range orders {
		if order.Status != "FILLED" {
			return false, nil
		}
	}
	return true, nil
}

func retryPendingOrders() error {
	storedOrders, err := loadStoredOrders()
	if err != nil {
		return fmt.Errorf("load stored orders: %w", err)
	}

	for i := range storedOrders {
		if storedOrders[i].Status == "PENDING_SUBMISSION" {

			log.Printf("Retrying pending order %s", storedOrders[i].ClientOrderID)

			order, err := createOrder(CreateOrderRequest{
				ClientOrderID: storedOrders[i].ClientOrderID,
				ProductID:     storedOrders[i].ProductID,
				Side:          storedOrders[i].Side,
				OrderConfiguration: OrderConfiguration{
					LimitLimitGTC: LimitLimitGTC{
						BaseSize:   storedOrders[i].BaseSize,
						LimitPrice: storedOrders[i].LimitPrice,
						PostOnly:   false,
					},
				},
			})
			if err != nil {
				log.Printf("ERROR: Failed to retry order %s: %v", storedOrders[i].ClientOrderID, err)
				return fmt.Errorf("retry order %s: %w", storedOrders[i].ClientOrderID, err)
			}
			storedOrders[i].OrderID = order.SuccessResponse.OrderID
			storedOrders[i].Status = "OPEN"
			// Once order is successfully opened, we can save the order details to the storedOrders slice and persist it to the JSON file
			if err := saveStoredOrders(storedOrders); err != nil {
				return fmt.Errorf("save pending order: %s: %w", storedOrders[i].ClientOrderID, err)
			}
			log.Printf("Recovered pending order %s, Coinbase order ID %s", storedOrders[i].ClientOrderID, storedOrders[i].ClientOrderID)
		}

	}
	return nil
}

func canCreatenewBatchOrders() (bool, error) {
	// an order could already be filled at Coinbase while its saved JSON still says OPEN
	if err := refreshStoredOrders(); err != nil {
		return false, fmt.Errorf("ERROR: refreshing stored orders: %w", err)
	}

	orders, err := loadStoredOrders()
	if err != nil {
		return false, err
	}

	// No previous orders so we can create new batch orders
	if len(orders) == 0 {
		return true, nil
	}

	// Only create new batch orders if all previous orders are filled
	for _, order := range orders {
		if order.Status != "FILLED" {
			return false, nil
		}
	}
	return true, nil
}
