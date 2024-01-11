package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/SiddyP/gotibber/tibber"
)

func queryUserExample() {

	ctx := context.Background()

	t := tibber.Client{
		APIClient: tibber.NewAPIClient(&tibber.APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		Logger: slog.Default(),
	}

	u := t.QueryUser(ctx, &tibber.User{})

	fmt.Printf("User: %v\n", u.Viewer.Name)
}
