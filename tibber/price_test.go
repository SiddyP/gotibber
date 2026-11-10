package tibber

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
	Auth      string         `json:"-"`
}

// newGQLServer serves body as the response to every GraphQL request and
// records the last request it received.
func newGQLServer(t *testing.T, body string) (*Client, *gqlRequest) {
	t.Helper()
	var got gqlRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		got.Auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	c := &Client{
		APIClient: NewAPIClient(&APIConfig{Token: "dummy_token", URL: srv.URL}),
		Logger:    slog.Default(),
	}
	return c, &got
}

func TestQueryHomePrice(t *testing.T) {
	c, got := newGQLServer(t, `{"data":{"viewer":{"home":{
		"id":"home-1",
		"currentSubscription":{"status":"running","priceInfo":{
			"current":{"total":1.5,"energy":1.0,"tax":0.5,"startsAt":"2026-11-10T13:15:00.000+01:00"},
			"today":[
				{"total":1.0,"energy":0.7,"tax":0.3,"startsAt":"2026-11-10T00:00:00.000+01:00"},
				{"total":2.0,"energy":1.4,"tax":0.6,"startsAt":"2026-11-10T00:15:00.000+01:00"}
			],
			"tomorrow":[
				{"total":3.0,"energy":2.1,"tax":0.9,"startsAt":"2026-11-11T00:00:00.000+01:00"}
			]
		}}
	}}}}`)

	info, err := c.QueryHomePrice(context.Background(), &HomePrice{
		HomeId:     "home-1",
		Resolution: PriceInfoResolutionQuarterHourly,
	})
	require.NoError(t, err)

	assert.Equal(t, "Bearer dummy_token", got.Auth)
	assert.Contains(t, got.Query, "home(id: $homeId)")
	assert.Equal(t, "home-1", got.Variables["homeId"])
	assert.Equal(t, "QUARTER_HOURLY", got.Variables["resolution"])

	require.Len(t, info.Today, 2)
	require.Len(t, info.Tomorrow, 1)
	assert.Equal(t, 2.0, info.Today[1].Total)
	assert.Equal(t, 3.0, info.Tomorrow[0].Total)
	assert.True(t, info.Today[1].StartsAt.Equal(time.Date(2026, 11, 9, 23, 15, 0, 0, time.UTC)))
	assert.Equal(t, 1.5, info.Current.Total)
}

func TestQueryHomePriceNoSubscription(t *testing.T) {
	c, _ := newGQLServer(t, `{"data":{"viewer":{"home":{"id":"home-1","currentSubscription":null}}}}`)

	_, err := c.QueryHomePrice(context.Background(), &HomePrice{HomeId: "home-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no current subscription")
}

func TestQueryHomePriceAPIError(t *testing.T) {
	c, _ := newGQLServer(t, `{"data":null,"errors":[{"message":"invalid home id"}]}`)

	_, err := c.QueryHomePrice(context.Background(), &HomePrice{HomeId: "home-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid home id")
}

func TestQueryHomePriceRequiresHomeId(t *testing.T) {
	c, _ := newGQLServer(t, `{}`)

	_, err := c.QueryHomePrice(context.Background(), &HomePrice{})
	assert.Error(t, err)
}
