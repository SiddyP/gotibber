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

// PricePoint is the price for one interval (an hour or a quarter-hour,
// depending on the requested PriceInfoResolution).
type PricePoint struct {
	StartsAt time.Time `json:"startsAt"`
	Total    float64   `json:"total"`
	Energy   float64   `json:"energy"`
	Tax      float64   `json:"tax"`
}

type PriceInfo struct {
	Today    []PricePoint `json:"today"`
	Tomorrow []PricePoint `json:"tomorrow"`
	Current  PricePoint   `json:"current"`
}

type PriceResponse struct {
	Viewer struct {
		Homes []struct {
			CurrentSubscription struct {
				PriceInfo PriceInfo `json:"priceInfo"`
			} `json:"currentSubscription"`
		} `json:"homes"`
	} `json:"viewer"`
}

// homePriceResponse is the raw shape of the HomePrice query. CurrentSubscription
// is a pointer so a home without a subscription (null) can be told apart from
// one whose price lists happen to be empty.
type homePriceResponse struct {
	Viewer struct {
		Home struct {
			Id                  string `json:"id"`
			CurrentSubscription *struct {
				Status    string    `json:"status"`
				PriceInfo PriceInfo `json:"priceInfo"`
			} `json:"currentSubscription"`
		} `json:"home"`
	} `json:"viewer"`
}
