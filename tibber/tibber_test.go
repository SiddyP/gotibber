package tibber

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type mockAPIClient struct {
	// runFn is the function that will be called when the Run method is called.
	runFn func(ctx context.Context, req *graphql.Request, resp any) error
}

func (m *mockAPIClient) Run(ctx context.Context, req *graphql.Request, resp any) error {
	return m.runFn(ctx, req, resp)
}

func TestClient_QueryUser(t *testing.T) {
	type fields struct {
		APIClient       APIClienter
		DBConn          *pgx.Conn
		Logger          *slog.Logger
		WebsocketClient *WebsocketClient
		Wg              *sync.WaitGroup
	}
	type args struct {
		ctx context.Context
		u   *User
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UserResponse
	}{
		{
			name: "simple ok query single value",
			fields: fields{
				APIClient: &mockAPIClient{
					runFn: func(_ context.Context, _ *graphql.Request, resp any) error {
						u := resp.(*UserResponse)
						u.Viewer.Name = "test"
						return nil
					},
				},
			},
			want: UserResponse{
				Viewer: struct {
					Name string "json:\"name\""
				}{
					Name: "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Client{
				APIClient:       tt.fields.APIClient,
				DBConn:          tt.fields.DBConn,
				Logger:          tt.fields.Logger,
				WebsocketClient: tt.fields.WebsocketClient,
				Wg:              tt.fields.Wg,
			}

			got := tr.QueryUser(tt.args.ctx, tt.args.u)
			require.NotNil(t, got)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestClient_QueryConsumption(t *testing.T) {
	type fields struct {
		APIClient       APIClienter
		DBConn          *pgx.Conn
		Logger          *slog.Logger
		WebsocketClient *WebsocketClient
		Wg              *sync.WaitGroup
	}
	type args struct {
		ctx context.Context
		c   *Consumption
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   HomeConsumptionResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Client{
				APIClient:       tt.fields.APIClient,
				DBConn:          tt.fields.DBConn,
				Logger:          tt.fields.Logger,
				WebsocketClient: tt.fields.WebsocketClient,
				Wg:              tt.fields.Wg,
			}
			if got := tr.QueryConsumption(tt.args.ctx, tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.QueryConsumption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_QueryPrice(t *testing.T) {
	type fields struct {
		APIClient       APIClienter
		DBConn          *pgx.Conn
		Logger          *slog.Logger
		WebsocketClient *WebsocketClient
		Wg              *sync.WaitGroup
	}
	type args struct {
		ctx context.Context
		p   *Price
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   PriceResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Client{
				APIClient:       tt.fields.APIClient,
				DBConn:          tt.fields.DBConn,
				Logger:          tt.fields.Logger,
				WebsocketClient: tt.fields.WebsocketClient,
				Wg:              tt.fields.Wg,
			}
			if got := tr.QueryPrice(tt.args.ctx, tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.QueryPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Subscribe(t *testing.T) {
	type fields struct {
		APIClient       APIClienter
		DBConn          *pgx.Conn
		Logger          *slog.Logger
		WebsocketClient *WebsocketClient
		Wg              *sync.WaitGroup
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Client{
				APIClient:       tt.fields.APIClient,
				DBConn:          tt.fields.DBConn,
				Logger:          tt.fields.Logger,
				WebsocketClient: tt.fields.WebsocketClient,
				Wg:              tt.fields.Wg,
			}
			tr.Subscribe(tt.args.ctx)
		})
	}
}
