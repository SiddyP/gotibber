package tibber

import "time"

type UserResponse struct {
	Viewer struct {
		Name string `json:"name"`
	}
}

type WebsocketSubscriptionUrlResponse struct {
	Viewer struct {
		Url string `json:"websocketSubscriptionUrl"`
	} `json:"viewer"`
}

type HomeConsumptionResponse struct {
	Viewer struct {
		Homes []struct {
			Consumption struct {
				Nodes []struct {
					From            time.Time `json:"from"`
					To              time.Time `json:"to"`
					ConsumptionUnit string    `json:"consumptionUnit"`
					Cost            float64   `json:"cost"`
					UnitPrice       float64   `json:"unitPrice"`
					UnitPriceVAT    float64   `json:"unitPriceVAT"`
					Consumption     float64   `json:"consumption"`
				} `json:"nodes"`
			} `json:"consumption"`
		} `json:"homes"`
	} `json:"viewer"`
}

type PriceResponse struct {
	Viewer struct {
		Homes []struct {
			CurrentSubscription struct {
				PriceInfo struct {
					Today []struct {
						StartsAt time.Time `json:"startsAt"`
						Total    float64   `json:"total"`
						Energy   float64   `json:"energy"`
						Tax      float64   `json:"tax"`
					} `json:"today"`
					Tomorrow []struct {
						StartsAt time.Time `json:"startsAt"`
						Total    float64   `json:"total"`
						Energy   float64   `json:"energy"`
						Tax      float64   `json:"tax"`
					} `json:"tomorrow"`
					Current struct {
						StartsAt time.Time `json:"startsAt"`
						Total    float64   `json:"total"`
						Energy   float64   `json:"energy"`
						Tax      float64   `json:"tax"`
					} `json:"current"`
				} `json:"priceInfo"`
			} `json:"currentSubscription"`
		} `json:"homes"`
	} `json:"viewer"`
}
