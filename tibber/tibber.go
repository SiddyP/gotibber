package tibber

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Client struct {
	APIClient       APIClienter
	DBConn          *pgx.Conn
	Logger          *slog.Logger
	WebsocketClient *WebsocketClient
	Wg              *sync.WaitGroup
}

type SubscriptionId struct{}

// Query Tibber API
func (t *Client) QueryUser(ctx context.Context, u *User) UserResponse {
	return u.query(ctx, t)
}

func (t *Client) QueryConsumption(ctx context.Context, c *Consumption) HomeConsumptionResponse {
	return c.query(ctx, t)
}

// func (t *Client) QueryWebsocketSubscriptionUrl(ctx context.Context, w *WebsocketSubscriptionUrl) WebsocketSubscriptionUrlResponse {
// 	return w.query(ctx, t)
// }

func (t *Client) QueryPrice(ctx context.Context, p *Price) PriceResponse {
	return p.query(ctx, t)
}

// Subscribe to Tibber websocket
func (t *Client) Subscribe(ctx context.Context) {
	defer t.Wg.Done()
	sctx := context.WithValue(ctx, SubscriptionId{}, uuid.New().String())
	socketConnection(sctx, t)
}
