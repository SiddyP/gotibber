package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SiddyP/gotibber/tibber"
)

func queryPriceExample() {

	ctx := context.Background()

	t := tibber.Client{
		APIClient: tibber.NewAPIClient(&tibber.APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		Logger: slog.Default(),
	}

	p := t.QueryPrice(ctx, &tibber.Price{})

	fmt.Printf("Price: %v\n", p.Viewer.Homes)
}
