package tibber

import "time"

type HomeConsumptionResponse struct {
	Viewer struct {
		Home struct {
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
		} `json:"home"`
	} `json:"viewer"`
}
