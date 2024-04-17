package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SiddyP/gotibber/tibber"
)

func main() {

	ctx := context.Background()

	t := tibber.NewClient(
		tibber.NewAPIClient(&tibber.APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		slog.Default(),
		nil,
		nil,
	)

	// bypasses the wrapping
	fmt.Println(t.QueryUser(ctx, &tibber.User{}))

	// goes via the querier
	tS := tibber.NewQuerier(t)
	fmt.Println(tS.QueryUser(ctx, &tibber.User{}))
	// fmt.Println(t.QueryUser(ctx, &tibber.User{}))
}
