package main

import (
	"net/http"
)

/*
JSON body to send for creating an order

	{
	  "client_order_id": "0000-00000-000000",
	  "product_id": "BTC-USD",
	  "side": "",
	  "order_configuration": {
	    "market_market_ioc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "rfq_disabled": true
	    },
	    "market_market_fok": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "rfq_disabled": true
	    },
	    "sor_limit_ioc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "rfq_disabled": true
	    },
	    "limit_limit_gtc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "post_only": false,
	      "rfq_disabled": true
	    },
	    "limit_limit_gtd": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "post_only": false
	    },
	    "limit_limit_fok": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "rfq_disabled": true
	    },
	    "twap_limit_gtd": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "start_time": "2021-05-31T07:59:59.000Z",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "limit_price": "10000.00",
	      "number_buckets": "5",
	      "bucket_size": "2.00",
	      "bucket_duration": "300s"
	    },
	    "stop_limit_stop_limit_gtc": {
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "stop_price": "20000.00",
	      "stop_direction": ""
	    },
	    "stop_limit_stop_limit_gtd": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_price": "20000.00",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "stop_direction": ""
	    },
	    "trigger_bracket_gtc": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_trigger_price": "20000.00"
	    },
	    "trigger_bracket_gtd": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_trigger_price": "20000.00",
	      "end_time": "2021-05-31T09:59:59.000Z"
	    },
	    "scaled_limit_gtc": {
	      "orders": [
	        {
	          "quote_size": "10.00",
	          "base_size": "0.001",
	          "limit_price": "10000.00",
	          "post_only": false,
	          "rfq_disabled": true
	        }
	      ],
	      "quote_size": "<string>",
	      "base_size": "<string>",
	      "num_orders": 123,
	      "min_price": "<string>",
	      "max_price": "<string>",
	      "price_distribution": "FLAT",
	      "size_distribution": "UNKNOWN_DISTRIBUTION",
	      "size_diff": "<string>",
	      "size_ratio": "<string>"
	    }
	  },
	  "leverage": "2.0",
	  "margin_type": "",
	  "retail_portfolio_id": "11111111-1111-1111-1111-111111111111",
	  "preview_id": "b40bbff9-17ce-4726-8b64-9de7ae57ad26",
	  "attached_order_configuration": {
	    "market_market_ioc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "rfq_disabled": true
	    },
	    "market_market_fok": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "rfq_disabled": true
	    },
	    "sor_limit_ioc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "rfq_disabled": true
	    },
	    "limit_limit_gtc": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "post_only": false,
	      "rfq_disabled": true
	    },
	    "limit_limit_gtd": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "post_only": false
	    },
	    "limit_limit_fok": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "rfq_disabled": true
	    },
	    "twap_limit_gtd": {
	      "quote_size": "10.00",
	      "base_size": "0.001",
	      "start_time": "2021-05-31T07:59:59.000Z",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "limit_price": "10000.00",
	      "number_buckets": "5",
	      "bucket_size": "2.00",
	      "bucket_duration": "300s"
	    },
	    "stop_limit_stop_limit_gtc": {
	      "base_size": "0.001",
	      "limit_price": "10000.00",
	      "stop_price": "20000.00",
	      "stop_direction": ""
	    },
	    "stop_limit_stop_limit_gtd": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_price": "20000.00",
	      "end_time": "2021-05-31T09:59:59.000Z",
	      "stop_direction": ""
	    },
	    "trigger_bracket_gtc": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_trigger_price": "20000.00"
	    },
	    "trigger_bracket_gtd": {
	      "base_size": 0.001,
	      "limit_price": "10000.00",
	      "stop_trigger_price": "20000.00",
	      "end_time": "2021-05-31T09:59:59.000Z"
	    },
	    "scaled_limit_gtc": {
	      "orders": [
	        {
	          "quote_size": "10.00",
	          "base_size": "0.001",
	          "limit_price": "10000.00",
	          "post_only": false,
	          "rfq_disabled": true
	        }
	      ],
	      "quote_size": "<string>",
	      "base_size": "<string>",
	      "num_orders": 123,
	      "min_price": "<string>",
	      "max_price": "<string>",
	      "price_distribution": "FLAT",
	      "size_distribution": "UNKNOWN_DISTRIBUTION",
	      "size_diff": "<string>",
	      "size_ratio": "<string>"
	    }
	  },
	  "sor_preference": "SOR_PREFERENCE_UNSPECIFIED",
	  "prediction_metadata": {
	    "prediction_side": "PREDICTION_SIDE_UNKNOWN",
	    "preview_order_est_average_filled_price": "<string>",
	    "supports_fractional_base_size": true
	  },
	  "cost_basis_method": "COST_BASIS_METHOD_UNSPECIFIED"
	}
*/ //TODO create payload from above json with structs
func createOrder() {

	req, _ := http.NewRequest("POST", "https://api.coinbase.com/api/v3/brokerage/orders", payload)
	req.Header.Set("Authorization", "Bearer "+getJwt("POST", "api.coinbase.com", "/api/v3/brokerage/orders"))
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
