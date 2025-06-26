package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

type OrdersResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []interface{} `json:"results"`
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func fetchAllOrders(baseURL, organizer, event, apiKey string) ([]interface{}, error) {
	var allOrders []interface{}
	url := fmt.Sprintf("%s/api/v1/organizers/%s/events/%s/orders/", baseURL, organizer, event)

	client := &http.Client{}

	for url != "" {
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
		fmt.Println(style.Render(fmt.Sprintf("Requesting orders: GET %s", url)))
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Token "+apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("bad status: %s\nBody: %s", resp.Status, string(body))
		}

		var ordersResp OrdersResponse
		if err := json.NewDecoder(resp.Body).Decode(&ordersResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		allOrders = append(allOrders, ordersResp.Results...)
		url = ordersResp.Next
	}

	return allOrders, nil
}

func main() {
	var baseURL, organizer, event, apiKey string
	var ok bool
	missing := false

 	if baseURL, ok = os.LookupEnv("PRETIX_BASEURL"); !ok || baseURL == "" {
 		fmt.Fprintln(os.Stderr, "Missing environment variable: PRETIX_BASEURL")
 		missing = true
 	}
 	if organizer, ok = os.LookupEnv("PRETIX_ORGANIZER"); !ok || organizer == "" {
 		fmt.Fprintln(os.Stderr, "Missing environment variable: PRETIX_ORGANIZER")
 		missing = true
 	}
 	if event, ok = os.LookupEnv("PRETIX_EVENT"); !ok || event == "" {
 		fmt.Fprintln(os.Stderr, "Missing environment variable: PRETIX_EVENT")
 		missing = true
 	}
 	if apiKey, ok = os.LookupEnv("PRETIX_APIKEY"); !ok || apiKey == "" {
 		fmt.Fprintln(os.Stderr, "Missing environment variable: PRETIX_APIKEY")
 		missing = true
 	}
 	if missing {
 		os.Exit(1)
 	}
 
	orders, err := fetchAllOrders(baseURL, organizer, event, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching orders: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Number of orders received: %d\n", len(orders))
}
