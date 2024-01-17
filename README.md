# gotibber
Query the tibber GraphQL API in a request/response fashion or setup a websocket connection to consume live measurements.

Websocket/streaming data requires a meter like the Tibber Pulse or Watty connected to the serial port (P1-port) of your powermeter. This repo uses the `graphql-transport-ws` sub-protocol for handling the subscription. 

## env
Provide your tibber details trough e.g. environment variables 
```zsh
export TIBBER_API_TOKEN=<REDACTED>
export TIBBER_HOUSE_ID=<REDACTED>
```

## example usage
Verify functionality by setting aforementioned environment variables and then run `go run .` in the `examples/`-directory

### query user
```go
func QueryUserExample() {

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
```

### query consumption

```go
func QueryConsumptionExample() {
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
```

### setup websocket

```go
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
	fmt.Println("*********************************\nShutdown signal received\n*********************************")
	cancelFunc()
	t.Wg.Wait()
	fmt.Println("All done!")
}

```

yields e.g.

```shell
New measurement: 2024-01-17 22:21:35 +0100 CET, 2834 W
New measurement: 2024-01-17 22:21:40 +0100 CET, 2835 W
New measurement: 2024-01-17 22:21:45 +0100 CET, 2839 W
New measurement: 2024-01-17 22:21:50 +0100 CET, 2841 W
New measurement: 2024-01-17 22:21:55 +0100 CET, 2842 W
```