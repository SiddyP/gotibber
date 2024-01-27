package tibber

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTibberClient(t *testing.T) {
	assert := assert.New(t)

	tC := Client{
		APIClient: NewAPIClient(&APIConfig{
			Token: "dummy_api_token",
			URL:   "https://api.tibber.com/v1-beta/gql",
		}),
		Logger: slog.Default(),
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
	assert.Equal(tC.APIClient.Config.Token, "dummy_api_token", "The two tokens should be the same.")
	assert.Equal(tC.WebsocketClient.Config.Token, "dummy_wss_api_token", "The two tokens should be the same.")
}
