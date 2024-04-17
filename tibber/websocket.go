package tibber

import (
	"context"
	"encoding/json"
	"fmt"
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

func connection_init(t *Client, c *websocket.Conn) error {
	m := InitMessage{
		Type: "connection_init",
		Payload: Payload{
			Token: t.WebsocketClient.Config.Token,
		},
	}
	err := c.WriteJSON(m)
	if err != nil {
		t.Logger.Error("connection_init", "error", err)
		return err
	}
	return nil
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
	err = c.WriteJSON(s)
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

func socketConnection(sctx context.Context, t *Client) {
	logger := t.Logger

	u := url.URL{Scheme: "wss", Host: t.WebsocketClient.Config.Host, Path: t.WebsocketClient.Config.Path}

	logger.Info("connecting to", "url", u.String())

	h := http.Header{}
	h.Set("Authorization", "Bearer "+t.WebsocketClient.Config.Token)
	h.Set("User-Agent", "REST github.com/SiddyP/gotibber/v0.3.0")
	h.Set("Sec-Websocket-Protocol", "graphql-transport-ws")
	h.Set("Accept-Encoding", "gzip, deflate, br")

	c, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	if err != nil {
		logger.Error("dial:", "error", err)
	}
	defer c.Close()

	err = connection_init(t, c)
	if err != nil {
		logger.Error("connection_init", "error", err)
	}

	// listener
	t.Wg.Add(1)
	go func() {
		defer t.Wg.Done()
		for {
			select {
			case <-sctx.Done():
				logger.Info("ctx.Done() shutting down listener")
				close(t.WebsocketClient.Data)
				return
			default:
				var m map[string]interface{}
				err = c.ReadJSON(&m)
				if err != nil {
					logger.Error("c.ReadJSON", "error", err)
				}
				switch msgType := m["type"]; msgType {
				case "connection_ack":
					err := subscribe(sctx, t, c)
					if err != nil {
						logger.Error("subscribe", "error", err)
					}
				case "next":
					streamingResp := StreamingQueryResponse{}

					jsonData, err := json.Marshal(m)
					if err != nil {
						fmt.Println("Error:", err)
					}

					err = json.Unmarshal(jsonData, &streamingResp)
					if err != nil {
						fmt.Println("Error:", err)
					}
					t.WebsocketClient.Data <- streamingResp.Payload.Data.LiveMeasurement
				}
			}
		}
	}()

	for {
		select {
		case <-sctx.Done():
			completedMessage := NewCompletedMessage(sctx.Value(SubscriptionId{}).(string))
			err = c.WriteJSON(completedMessage)
			if err != nil {
				logger.Error("<-interrupt", "error", err)
			}
			logger.Info("ctx.Done()! closing channel and and websocket.", "completedMessage", completedMessage)
			return
		}
	}
}
