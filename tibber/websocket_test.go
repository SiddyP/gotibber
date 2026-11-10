package tibber

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTibber is a minimal graphql-transport-ws server. For every accepted
// connection it performs the connection_init/ack + subscribe handshake and
// then calls handle with the 1-based connection attempt number and the
// subscription id. The connection is closed when handle returns.
type fakeTibber struct {
	srv      *httptest.Server
	attempts atomic.Int32
}

func newFakeTibber(t *testing.T, reject func(attempt int32) bool, handle func(attempt int32, c *websocket.Conn, subID string)) *fakeTibber {
	t.Helper()
	f := &fakeTibber{}
	upgrader := websocket.Upgrader{Subprotocols: []string{"graphql-transport-ws"}}
	f.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := f.attempts.Add(1)
		if reject != nil && reject(n) {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		var m serverMessage
		if err := c.ReadJSON(&m); err != nil || m.Type != "connection_init" {
			return
		}
		if err := c.WriteJSON(map[string]string{"type": "connection_ack"}); err != nil {
			return
		}
		if err := c.ReadJSON(&m); err != nil || m.Type != "subscribe" {
			return
		}
		handle(n, c, m.Id)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func sendMeasurement(c *websocket.Conn, subID string, power float64) error {
	return c.WriteJSON(map[string]any{
		"id":   subID,
		"type": "next",
		"payload": map[string]any{"data": map[string]any{"liveMeasurement": map[string]any{
			"timestamp": "2026-11-10T13:00:00+01:00",
			"power":     power,
		}}},
	})
}

// serverRestart closes the connection the way Tibber does on a deploy.
func serverRestart(c *websocket.Conn) {
	_ = c.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseGoingAway, "Server restart"),
		time.Now().Add(time.Second))
}

// drain blocks until the client closes the connection.
func drain(c *websocket.Conn) {
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}

func (f *fakeTibber) client(t *testing.T) *Client {
	t.Helper()
	u, err := url.Parse(f.srv.URL)
	require.NoError(t, err)
	return &Client{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		WebsocketClient: &WebsocketClient{
			Config: NewWebsocketConfig(&WebsocketConfig{
				Token: "dummy_wss_api_token",
				Host:  u.Host,
				Path:  "/v1-beta/gql/subscriptions",
				Id:    "dummy_house_id",
				Dialer: &websocket.Dialer{
					TLSClientConfig:  f.srv.Client().Transport.(*http.Transport).TLSClientConfig,
					HandshakeTimeout: time.Second,
				},
				ReadTimeout: 300 * time.Millisecond,
				MinBackoff:  10 * time.Millisecond,
				MaxBackoff:  40 * time.Millisecond,
			}),
			Data: make(chan LiveMeasurement),
		},
		Wg: &sync.WaitGroup{},
	}
}

// start runs Subscribe and registers a cleanup that cancels it and asserts
// that it shuts down and closes the Data channel.
func start(t *testing.T, c *Client) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	c.Wg.Add(1)
	go c.Subscribe(ctx)
	t.Cleanup(func() {
		cancel()
		done := make(chan struct{})
		go func() { c.Wg.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("Subscribe did not return after cancel")
		}
		_, ok := <-c.WebsocketClient.Data
		assert.False(t, ok, "Data channel should be closed after shutdown")
	})
	return cancel
}

func recv(t *testing.T, c *Client) LiveMeasurement {
	t.Helper()
	select {
	case m, ok := <-c.WebsocketClient.Data:
		require.True(t, ok, "Data channel closed unexpectedly")
		return m
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for a liveMeasurement")
	}
	return LiveMeasurement{}
}

// Regression test: a "close 1001 (going away): Server restart" from Tibber used
// to leave the client re-reading the dead socket until gorilla/websocket
// panicked with "repeated read on failed websocket connection".
func TestSubscribeReconnectsAfterServerRestart(t *testing.T) {
	f := newFakeTibber(t, nil, func(n int32, c *websocket.Conn, subID string) {
		if err := sendMeasurement(c, subID, float64(n)); err != nil {
			return
		}
		serverRestart(c)
	})
	c := f.client(t)
	start(t, c)

	for want := 1.0; want <= 3; want++ {
		m := recv(t, c)
		require.NotNil(t, m.Power)
		assert.Equal(t, want, *m.Power)
	}
}

func TestSubscribeReconnectsAfterReadTimeout(t *testing.T) {
	f := newFakeTibber(t, nil, func(n int32, c *websocket.Conn, subID string) {
		if n == 1 {
			drain(c) // silent, half-dead connection
			return
		}
		if err := sendMeasurement(c, subID, 2); err == nil {
			drain(c)
		}
	})
	c := f.client(t)
	start(t, c)

	m := recv(t, c)
	assert.Equal(t, 2.0, *m.Power)
	assert.GreaterOrEqual(t, f.attempts.Load(), int32(2))
}

func TestSubscribeRetriesFailedDial(t *testing.T) {
	f := newFakeTibber(t,
		func(n int32) bool { return n <= 2 },
		func(n int32, c *websocket.Conn, subID string) {
			if err := sendMeasurement(c, subID, float64(n)); err == nil {
				drain(c)
			}
		})
	c := f.client(t)
	start(t, c)

	m := recv(t, c)
	assert.Equal(t, 3.0, *m.Power)
	assert.Equal(t, int32(3), f.attempts.Load())
}

func TestSubscribeAnswersPing(t *testing.T) {
	f := newFakeTibber(t, nil, func(n int32, c *websocket.Conn, subID string) {
		if err := c.WriteJSON(map[string]string{"type": "ping"}); err != nil {
			return
		}
		var m serverMessage
		if err := c.ReadJSON(&m); err != nil || m.Type != "pong" {
			return
		}
		if err := sendMeasurement(c, subID, 42); err == nil {
			drain(c)
		}
	})
	c := f.client(t)
	start(t, c)

	m := recv(t, c)
	assert.Equal(t, 42.0, *m.Power)
	assert.Equal(t, int32(1), f.attempts.Load(), "pong should keep the first connection alive")
}

func TestSubscribeStopsWhenConsumerIsGone(t *testing.T) {
	f := newFakeTibber(t, nil, func(n int32, c *websocket.Conn, subID string) {
		for {
			if err := sendMeasurement(c, subID, 1); err != nil {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	})
	c := f.client(t)
	cancel := start(t, c)

	recv(t, c)
	// Stop consuming and cancel: Subscribe must not block forever trying to
	// hand over the next measurement. The cleanup registered by start asserts
	// that it returns.
	cancel()
}
