package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/SiddyP/gotibber/tibber"
)

func setupWebsocket() {
	// terminate listens for SIGINT and SIGTERM signals from the OS
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancelFunc := context.WithCancel(context.Background())

	t := tibber.Client{
		APIClient: tibber.NewAPIClient(&tibber.APIConfig{
			Token: os.Getenv("TIBBER_API_TOKEN"),
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		Logger: slog.Default(),
		WebsocketClient: &tibber.WebsocketClient{
			Config: tibber.NewWebsocketConfig(&tibber.WebsocketConfig{
				Token: os.Getenv("TIBBER_API_TOKEN"),
				Host:  "websocket-api.tibber.com",
				Path:  "/v1-beta/gql/subscriptions",
				Id:    os.Getenv("TIBBER_HOUSE_ID"),
			}),
			Data: make(chan tibber.LiveMeasurement),
		},
		Wg: &sync.WaitGroup{},
	}

	t.Wg.Add(1)
	go t.Subscribe(ctx)

	go func() {
		for {
			select {
			case liveMeasurement := <-t.WebsocketClient.Data:
				fmt.Printf(
					"New measurement: %v, %v W\n",
					*liveMeasurement.Timestamp, *liveMeasurement.Power,
				)
			case <-ctx.Done():
				fmt.Println("Done!")
				return
			}
		}
	}()

	<-terminate //block until terminate is closed
	fmt.Println("*********************\nShutdown signal received\n*********************")
	cancelFunc()
	t.Wg.Wait()
	fmt.Println("All done!")
}
