package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func sendNotification(title, message string) error {
	token := getPushoverToken()
	user := getPushoverUser()

	if token == "" || user == "" {
		log.Printf("ERROR: Pushover token or user key is not set. Skipping notification.")
		return nil
	}

	form := url.Values{}
	form.Set("token", token)
	form.Set("user", user)
	form.Set("title", title)
	form.Set("message", message)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.PostForm(
		"https://api.pushover.net/1/messages.json",
		form,
	)
	if err != nil {
		return fmt.Errorf("sending Pushover notification: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Pushover returned %s: %s", resp.Status, body)
		return fmt.Errorf("Pushover returned %s: %s", resp.Status, body)
	}

	var result struct {
		Status int `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("reading Pushover response: %w", err)
	}
	if result.Status != 1 {
		log.Printf("ERROR: Pushover did not accept the message: %s", body)
		return fmt.Errorf("Pushover did not accept the message: %s", body)
	}

	return nil
}
