package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/coinbase-samples/advanced-trade-sdk-go/accounts"
	"github.com/coinbase-samples/advanced-trade-sdk-go/client"
	"github.com/coinbase-samples/advanced-trade-sdk-go/credentials"
)

func main() {
	var creds credentials.Credentials

	if err := json.Unmarshal([]byte(os.Getenv("ADVANCED_CREDENTIALS")), &creds); err != nil {
		log.Fatalf("unable to deserialize advanced trade credentials JSON: %v", err)
	}

	httpClient, err := client.DefaultHttpClient()
	if err != nil {
		log.Fatalf("unable to load default http client: %v", err)
	}

	restClient := client.NewRestClient(&creds, httpClient)
	//portfoliosService := portfolios.NewPortfoliosService(restClient)

	accountsService := accounts.NewAccountsService(restClient)

	resp, err := accountsService.ListAccounts(context.Background(), &accounts.ListAccountsRequest{})
	if err != nil {
		log.Fatalf("error listing accounts: %v", err)
	}
	//resp, err := portfoliosService.ListPortfolios(context.Background(), &portfolios.ListPortfoliosRequest{})
	//if err != nil {
	//	log.Fatalf("error listing portfolios: %v", err)
	//}

	jsonResponse, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("error marshaling response to JSON: %v", err))
	}
	fmt.Println(string(jsonResponse))
}
