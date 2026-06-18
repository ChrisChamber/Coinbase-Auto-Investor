package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Account struct {
	UUID             string `json:"uuid"`
	Currency         string `json:"currency"`
	Name             string `json:"name"`
	AvailableBalance struct {
		Currency string `json:"currency"`
		Value    string `json:"value"`
	} `json:"available_balance"`
}

type AccountsResp struct {
	Accounts []Account `json:"accounts"`
	Cursor   string    `json:"cursor"`
	HasNext  bool      `json:"has_next"`
	Size     int       `json:"size"`
}

func getAccounts() (AccountsResp, *Account, *Account, error) {
	req, _ := http.NewRequest("GET", "https://api.coinbase.com/api/v3/brokerage/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+getJwt("GET", "api.coinbase.com", "/api/v3/brokerage/accounts"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return AccountsResp{}, nil, nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccountsResp{}, nil, nil, fmt.Errorf("error reading response body: %w", err)
	}

	var accounts AccountsResp
	if err := json.Unmarshal(body, &accounts); err != nil {
		log.Fatalf("unmarshal accounts: %v", err)
	}
	var eur, btc *Account
	// scan current page first
	for i := range accounts.Accounts {
		a := &accounts.Accounts[i]
		if a.Currency == "EUR" && eur == nil {
			eur = a
		}
		if a.Currency == "BTC" && btc == nil {
			btc = a
		}
	}
	// if either account is not found, keep fetching next pages until found or no more pages
	for eur == nil || btc == nil {
		if !accounts.HasNext {
			break
		}
		req, _ := http.NewRequest("GET", "https://api.coinbase.com/api/v3/brokerage/accounts?cursor="+accounts.Cursor, nil)
		req.Header.Set("Authorization", "Bearer "+getJwt("GET", "api.coinbase.com", "/api/v3/brokerage/accounts"))
		req.Header.Set("Accept", "application/json")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return accounts, eur, btc, fmt.Errorf("error making request: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return accounts, eur, btc, fmt.Errorf("error reading response body: %w", err)
		}
		if err := json.Unmarshal(body, &accounts); err != nil {
			return accounts, eur, btc, fmt.Errorf("unmarshal accounts: %w", err)
		}

		for i := range accounts.Accounts {
			a := &accounts.Accounts[i]
			if a.Currency == "EUR" && eur == nil {
				eur = a
			}
			if a.Currency == "BTC" && btc == nil {
				btc = a
			}
		}
	}
	if eur != nil {
		fmt.Printf("EUR account: id=%s name=%s balance=%s %s\n", eur.UUID, eur.Name, eur.AvailableBalance.Value, eur.AvailableBalance.Currency)
	} else {
		fmt.Println("EUR account not found")

	}
	if btc != nil {
		fmt.Printf("BTC account: id=%s name=%s balance=%s %s\n", btc.UUID, btc.Name, btc.AvailableBalance.Value, btc.AvailableBalance.Currency)
	} else {
		fmt.Println("BTC account not found")
	}

	return accounts, eur, btc, nil
}
