package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SiddyP/gotibber/tibber"
)

func queryConsumptionExample() {
	ctx := context.Background()

	t := tibber.Client{
		APIClient: tibber.NewAPIClient(&tibber.APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		Logger: slog.Default(),
	}

	c := t.QueryConsumption(ctx, &tibber.Consumption{
		Id:         os.Getenv("TIBBER_HOUSE_ID"),
		Resolution: "HOURLY",
		Last:       5,
	})

	fmt.Printf("Consumption: %v\n", c.Viewer.Homes)
}
