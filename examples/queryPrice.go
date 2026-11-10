package main

import (
	"context"
	"fmt"
	"log"
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

	p, err := t.QueryHomePrice(ctx, &tibber.HomePrice{
		HomeId:     os.Getenv("TIBBER_HOUSE_ID"),
		Resolution: tibber.PriceInfoResolutionQuarterHourly,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current price: %v, today: %d prices, tomorrow: %d prices\n",
		p.Current.Total, len(p.Today), len(p.Tomorrow))
}
