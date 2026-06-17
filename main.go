package main

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
	"github.com/square/go-jose/v3"
	"github.com/square/go-jose/v3/jwt"
)

var (
	keyName       = os.Getenv("COINBASE_KEY_NAME")
	keySecret     = os.Getenv("COINBASE_KEY_SECRET")
	requestMethod = "GET"
	requestHost   = "api.coinbase.com"
	requestPath   = "/api/v3/brokerage/accounts"
)

type APIKeyClaims struct {
	*jwt.Claims
	URI string `json:"uri"`
}

func buildJWT(uri string) (string, error) {
	block, _ := pem.Decode([]byte(keySecret))
	if block == nil {
		return "", fmt.Errorf("jwt: Could not decode private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}

	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{NonceSource: nonceSource{}}).WithType("JWT").WithHeader("kid", keyName),
	)
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}

	cl := &APIKeyClaims{
		Claims: &jwt.Claims{
			Subject:   keyName,
			Issuer:    "cdp",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
		},
		URI: uri,
	}
	jwtString, err := jwt.Signed(sig).Claims(cl).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}
	return jwtString, nil
}

var max = big.NewInt(math.MaxInt64)

type nonceSource struct{}

func (n nonceSource) Nonce() (string, error) {
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

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

func getJwt() string {
	uri := fmt.Sprintf("%s %s%s", requestMethod, requestHost, requestPath)
	jwt, err := buildJWT(uri)
	if err != nil {
		log.Errorf("error building jwt: %v", err)
	}
	return jwt
}

func getAccounts() (AccountsResp, error) {

	req, _ := http.NewRequest("GET", "https://api.coinbase.com/api/v3/brokerage/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+getJwt())
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error reading response body: %v", err)
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
		req.Header.Set("Authorization", "Bearer "+getJwt())
		req.Header.Set("Accept", "application/json")
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			return accounts, fmt.Errorf("error making request: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return accounts, fmt.Errorf("error reading response body: %w", err)
		}
		if err := json.Unmarshal(body, &accounts); err != nil {
			return accounts, fmt.Errorf("unmarshal accounts: %w", err)
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

	return accounts, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	accounts, err := getAccounts()
	if err != nil {
		log.Fatalf("error getting accounts: %v", err)
	}
	fmt.Printf("Retrieved %d accounts on final page\n", accounts.Size)
}
