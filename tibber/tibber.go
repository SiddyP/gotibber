package tibber

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	APIClient       GQLClienter
	DBConn          *pgxpool.Pool
	Logger          *slog.Logger
	WebsocketClient *WebsocketClient
	Wg              *sync.WaitGroup
}

type SubscriptionId struct{}

func NewClient(c GQLClienter, l *slog.Logger, w *WebsocketClient, wg *sync.WaitGroup) *Client {
	return &Client{
		APIClient:       c,
		Logger:          l,
		WebsocketClient: w,
		Wg:              wg,
	}
}

// Query Tibber API
func (t *Client) QueryUser(ctx context.Context, u *User) UserResponse {
	return u.query(ctx, t)
}

func (t *Client) QueryConsumption(ctx context.Context, c *Consumption) HomeConsumptionResponse {
	return c.query(ctx, t)
}

func (t *Client) QueryWebsocketSubscriptionUrl(ctx context.Context, w *WebsocketSubscriptionUrl) WebsocketSubscriptionUrlResponse {
	return w.query(ctx, t)
}

// QueryPrice returns price info for every home on the account, in API order.
//
// Deprecated: the account can contain homes without a subscription (empty
// price info), and they may be listed first. Use QueryHomePrice instead.
func (t *Client) QueryPrice(ctx context.Context, p *Price) PriceResponse {
	return p.query(ctx, t)
}

// QueryHomePrice returns today's, tomorrow's and the current price for the
// home p.HomeId. It returns an error if the request fails or the home has no
// current subscription.
func (t *Client) QueryHomePrice(ctx context.Context, p *HomePrice) (PriceInfo, error) {
	return p.query(ctx, t)
}

// Subscribe to Tibber websocket
func (t *Client) Subscribe(ctx context.Context) {
	defer t.Wg.Done()
	sctx := context.WithValue(ctx, SubscriptionId{}, uuid.New().String())
	socketConnection(sctx, t)
}
