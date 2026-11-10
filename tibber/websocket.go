package tibber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type InitMessage struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	Token string `json:"token"`
}

type Subscribe struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Payload struct {
		Variables struct {
		} `json:"variables"`
		Extensions struct {
		} `json:"extensions"`
		Query string `json:"query"`
	} `json:"payload"`
}

type CompletedMessage struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type LiveMeasurement struct {
	Timestamp                      *time.Time `json:"timestamp"`
	Currency                       *string    `json:"currency"`
	LastMeterProduction            *float64   `json:"lastMeterProduction"`
	AveragePower                   *float64   `json:"averagePower"`
	MaxPowerProduction             *float64   `json:"maxPowerProduction"`
	AccumulatedReward              *float64   `json:"accumulatedReward"`
	MinPower                       *float64   `json:"minPower"`
	AccumulatedCost                *float64   `json:"accumulatedCost"`
	CurrentL1                      *float64   `json:"currentL1"`
	MinPowerProduction             *float64   `json:"minPowerProduction"`
	CurrentL3                      *float64   `json:"currentL3"`
	LastMeterConsumption           *float64   `json:"lastMeterConsumption"`
	AccumulatedConsumption         *float64   `json:"accumulatedConsumption"`
	MaxPower                       *float64   `json:"maxPower"`
	AccumulatedProductionLastHour  *float64   `json:"accumulatedProductionLastHour"`
	AccumulatedProduction          *float64   `json:"accumulatedProduction"`
	CurrentL2                      *float64   `json:"currentL2"`
	Power                          *float64   `json:"power"`
	PowerFactor                    *float64   `json:"powerFactor"`
	PowerProduction                *float64   `json:"powerProduction"`
	PowerProductionReactive        *float64   `json:"powerProductionReactive"`
	PowerReactive                  *float64   `json:"powerReactive"`
	SignalStrength                 *float64   `json:"signalStrength"`
	AccumulatedConsumptionLastHour *float64   `json:"accumulatedConsumptionLastHour"`
	VoltagePhase1                  *float64   `json:"voltagePhase1"`
	VoltagePhase2                  *float64   `json:"voltagePhase2"`
	VoltagePhase3                  *float64   `json:"voltagePhase3"`
}

type Data struct {
	LiveMeasurement LiveMeasurement `json:"liveMeasurement"`
}

type ResponsePayload struct {
	Data Data `json:"data"`
}

type StreamingQueryResponse struct {
	Payload ResponsePayload `json:"payload"`
	Id      string          `json:"id"`
	Type    string          `json:"type"`
}

func NewCompletedMessage(id string) CompletedMessage {
	return CompletedMessage{
		Id:   id,
		Type: "complete",
	}
}

const (
	userAgent = "REST github.com/SiddyP/gotibber/v0.6.0"

	defaultReadTimeout = 90 * time.Second
	defaultMinBackoff  = 5 * time.Second
	defaultMaxBackoff  = 5 * time.Minute
	writeTimeout       = 10 * time.Second
)

// serverMessage is a graphql-transport-ws message received from the server.
type serverMessage struct {
	Id      string          `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// nextPayload is the payload of a "next" message.
type nextPayload struct {
	Data   Data `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func writeJSON(c *websocket.Conn, v any) error {
	if err := c.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	return c.WriteJSON(v)
}

func connection_init(t *Client, c *websocket.Conn) error {
	m := InitMessage{
		Type: "connection_init",
		Payload: Payload{
			Token: t.WebsocketClient.Config.Token,
		},
	}
	return writeJSON(c, m)
}

func subscribe(sctx context.Context, t *Client, c *websocket.Conn) error {
	data := fmt.Sprintf(
		`{
			"id":"%s",
			"type":"subscribe",
			"payload":{"variables":{},"extensions":{},"query":"subscription { liveMeasurement(homeId: \"%s\") {  timestamp power lastMeterConsumption accumulatedConsumption accumulatedProduction accumulatedConsumptionLastHour accumulatedProductionLastHour accumulatedCost accumulatedReward currency minPower averagePower maxPower powerProduction powerReactive powerProductionReactive minPowerProduction maxPowerProduction lastMeterProduction powerFactor voltagePhase1 voltagePhase2 voltagePhase3 currentL1 currentL2 currentL3 signalStrength } }"}
		}`, sctx.Value(SubscriptionId{}).(string), t.WebsocketClient.Config.Id)

	var s Subscribe
	err := json.Unmarshal([]byte(data), &s)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	t.Logger.Info("subscribing..", "gqlQuery", s)
	err = writeJSON(c, s)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	return nil
}

type WebsocketConfig struct {
	Token string
	Host  string
	Path  string
	Id    string

	// Optional. Zero values fall back to the defaults.
	Dialer      *websocket.Dialer // default websocket.DefaultDialer
	ReadTimeout time.Duration     // reconnect if nothing arrives for this long (default 90s)
	MinBackoff  time.Duration     // first reconnect delay (default 5s)
	MaxBackoff  time.Duration     // cap for the exponential reconnect delay (default 5m)
}

func (w *WebsocketConfig) dialer() *websocket.Dialer {
	if w.Dialer != nil {
		return w.Dialer
	}
	return websocket.DefaultDialer
}

func orDefault(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

func (w *WebsocketConfig) readTimeout() time.Duration {
	return orDefault(w.ReadTimeout, defaultReadTimeout)
}
func (w *WebsocketConfig) minBackoff() time.Duration {
	return orDefault(w.MinBackoff, defaultMinBackoff)
}
func (w *WebsocketConfig) maxBackoff() time.Duration {
	return orDefault(w.MaxBackoff, defaultMaxBackoff)
}

type WebsocketClient struct {
	Config *WebsocketConfig
	Data   chan LiveMeasurement
}

func NewWebsocketConfig(w *WebsocketConfig) *WebsocketConfig {
	if w.Token == "" {
		panic("TIBBER_API_TOKEN not set")
	}

	if w.Host == "" {
		panic("TIBBER_API_HOST not set")
	}

	if w.Path == "" {
		panic("TIBBER_API_PATH not set")
	}
	if w.Id == "" {
		panic("TIBBER_HOUSE_ID not set")
	}
	return w
}

// socketConnection keeps the liveMeasurement subscription alive until sctx is
// cancelled. Whenever a connection fails (dial error, server restart, read
// timeout, ...) it reconnects after an exponential, jittered backoff that
// resets once a connection has delivered data. It closes
// t.WebsocketClient.Data before returning.
func socketConnection(sctx context.Context, t *Client) {
	defer close(t.WebsocketClient.Data)
	cfg := t.WebsocketClient.Config

	backoff := cfg.minBackoff()
	for {
		gotData, err := runSubscription(sctx, t)
		if sctx.Err() != nil {
			t.Logger.Info("ctx.Done() websocket closed")
			return
		}
		if gotData {
			backoff = cfg.minBackoff()
		}
		// Jitter within [backoff/2, backoff] so clients don't all reconnect at
		// the same instant after a Tibber server restart.
		wait := backoff/2 + rand.N(backoff/2+1)
		t.Logger.Warn("websocket disconnected, reconnecting", "error", err, "retry_in", wait.String())
		select {
		case <-sctx.Done():
			t.Logger.Info("ctx.Done() websocket closed")
			return
		case <-time.After(wait):
		}
		backoff = min(2*backoff, cfg.maxBackoff())
	}
}

// runSubscription runs a single connection: dial, connection_init, subscribe
// on connection_ack, then forward liveMeasurements to t.WebsocketClient.Data
// until the connection fails or sctx is cancelled. It never reads from a
// connection after a read error. It reports whether any measurement was
// forwarded.
func runSubscription(sctx context.Context, t *Client) (gotData bool, err error) {
	logger := t.Logger
	cfg := t.WebsocketClient.Config

	u := url.URL{Scheme: "wss", Host: cfg.Host, Path: cfg.Path}
	h := http.Header{}
	h.Set("Authorization", "Bearer "+cfg.Token)
	h.Set("User-Agent", userAgent)
	h.Set("Sec-Websocket-Protocol", "graphql-transport-ws")
	h.Set("Accept-Encoding", "gzip, deflate, br")

	logger.Info("connecting to", "url", u.String())
	c, _, err := cfg.dialer().DialContext(sctx, u.String(), h)
	if err != nil {
		return false, fmt.Errorf("dial: %w", err)
	}
	defer c.Close()

	// Reads block; on cancellation send a close frame and close the
	// connection so the read loop below returns promptly. WriteControl and
	// Close are safe to call concurrently with the reader/writer.
	stop := context.AfterFunc(sctx, func() {
		_ = c.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second))
		c.Close()
	})
	defer stop()

	if err := connection_init(t, c); err != nil {
		return false, fmt.Errorf("connection_init: %w", err)
	}

	for {
		if err := c.SetReadDeadline(time.Now().Add(cfg.readTimeout())); err != nil {
			return gotData, fmt.Errorf("set read deadline: %w", err)
		}
		var msg serverMessage
		if err := c.ReadJSON(&msg); err != nil {
			return gotData, fmt.Errorf("read: %w", err)
		}

		switch msg.Type {
		case "connection_ack":
			if err := subscribe(sctx, t, c); err != nil {
				return gotData, fmt.Errorf("subscribe: %w", err)
			}
		case "ping":
			if err := writeJSON(c, map[string]string{"type": "pong"}); err != nil {
				return gotData, fmt.Errorf("pong: %w", err)
			}
		case "next":
			var p nextPayload
			if err := json.Unmarshal(msg.Payload, &p); err != nil {
				logger.Error("decode next payload", "error", err)
				continue
			}
			if len(p.Errors) > 0 {
				logger.Error("liveMeasurement returned errors", "errors", fmt.Sprint(p.Errors))
				continue
			}
			select {
			case t.WebsocketClient.Data <- p.Data.LiveMeasurement:
				gotData = true
			case <-sctx.Done():
				return gotData, sctx.Err()
			}
		case "error":
			return gotData, fmt.Errorf("subscription error: %s", msg.Payload)
		case "complete":
			return gotData, errors.New("server completed the subscription")
		}
	}
}
