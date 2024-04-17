package tibber

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

// MockClient is the mock client
type MockGQLClient struct {
	runFn func(ctx context.Context, req *graphql.Request, resp interface{}) error
}

func (m *MockGQLClient) Run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	fmt.Println("MockGQLClient.Run")
	return m.runFn(ctx, req, resp)
}

func TestCreateTibberClient(t *testing.T) {
	assert := assert.New(t)

	mClient := &MockGQLClient{
		runFn: func(ctx context.Context, req *graphql.Request, resp interface{}) error { return nil },
	}

	tC := Client{
		APIClient: mClient,
		Logger:    slog.Default(),
		WebsocketClient: &WebsocketClient{
			Config: NewWebsocketConfig(&WebsocketConfig{
				Token: "dummy_wss_api_token",
				Host:  "websocket-api.tibber.com",
				Path:  "/v1-beta/gql/subscriptions",
				Id:    "dummy_house_id",
			}),
			Data: make(chan LiveMeasurement),
		},
	}
	assert.Equal(mClient, tC.APIClient)
}

func TestQueryUser(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mClient := &MockGQLClient{
		runFn: func(ctx context.Context, req *graphql.Request, resp interface{}) error { return nil },
	}

	tC := Client{
		APIClient:       mClient,
		Logger:          slog.Default(),
		WebsocketClient: nil,
	}

	resp := UserResponse{
		Viewer: struct {
			Name string `json:"name"`
		}{
			Name: "",
		},
	}

	u := tC.QueryUser(ctx, &User{})

	assert.Equal(resp, u)
}

func TestQueryWebsocketSubscriptionUrl(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	mClient := &MockGQLClient{
		runFn: func(ctx context.Context, req *graphql.Request, resp interface{}) error { return nil },
	}

	tC := Client{
		APIClient:       mClient,
		Logger:          slog.Default(),
		WebsocketClient: nil,
	}

	resp := WebsocketSubscriptionUrlResponse{
		Viewer: struct {
			Url string `json:"websocketSubscriptionUrl"`
		}{
			Url: "",
		},
	}

	u := tC.QueryWebsocketSubscriptionUrl(ctx, &WebsocketSubscriptionUrl{})

	assert.Equal(resp, u)
}
