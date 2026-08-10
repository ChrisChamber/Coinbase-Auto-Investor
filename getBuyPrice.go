package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shopspring/decimal"
)

type PriceResponse struct {
	Data struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"data"`
}

func getBuyPrice(productID string) (decimal.Decimal, error) {
	req, _ := http.NewRequest("GET", "https://api.coinbase.com/v2/prices/"+productID+"/buy", nil)
	//req.Header.Set("Authorization", "Bearer "+getJwt("GET", "api.coinbase.com", "/v2/prices/"+productID+"/buy"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("error fetching buy price: %s", resp.Status)
	}
	// Parse the response body to extract the buy price
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error reading response body: %w", err)
	}
	var priceResponse PriceResponse
	if err := json.Unmarshal(respBody, &priceResponse); err != nil {
		return decimal.Zero, fmt.Errorf("error unmarshaling response: %w", err)
	}
	fmt.Printf("Current buy price for %s: %s\n", productID, priceResponse.Data.Amount)
	price, err := decimal.NewFromString(priceResponse.Data.Amount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error parsing buy price: %w", err)
	}
	return price, nil
}
