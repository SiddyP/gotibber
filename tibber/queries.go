package tibber

import (
	"context"
	"log"

	"github.com/machinebox/graphql"
)

type User struct {
}

type Consumption struct {
	Id         string
	Resolution string
	Last       int
}

type PriceInfoResolution string

const (
	PriceInfoResolutionHourly        PriceInfoResolution = "HOURLY"
	PriceInfoResolutionQuarterHourly PriceInfoResolution = "QUARTER_HOURLY"
)

type Price struct {
	Resolution PriceInfoResolution
}

type WebsocketSubscriptionUrl struct {
}

func (w *WebsocketSubscriptionUrl) query(ctx context.Context, t *Client) WebsocketSubscriptionUrlResponse {
	req := graphql.NewRequest(`
		query {
			viewer {
				websocketSubscriptionUrl
			}
		}
	`)

	var ws WebsocketSubscriptionUrlResponse
	if err := t.APIClient.Run(ctx, req, &ws); err != nil {
		log.Fatal(err)
	}
	return ws
}

func (q *User) query(ctx context.Context, t *Client) UserResponse {
	req := graphql.NewRequest(`
		query {
			viewer {
				name
			}
		}
	`)
	// run it and capture the response
	var u UserResponse
	if err := t.APIClient.Run(ctx, req, &u); err != nil {
		log.Fatal(err)
	}
	return u
}

func (q *Consumption) query(ctx context.Context, t *Client) HomeConsumptionResponse {
	req := graphql.NewRequest(`
			query ($resolution: EnergyResolution!, $last: Int!) {
				viewer {
					homes  {
						consumption(resolution: $resolution, last: $last) {
							nodes {
							from
							to
							cost
							unitPrice
							unitPriceVAT
							consumption
							consumptionUnit
							}
						}
					}
				}
			}
		`)

	// set any variables
	req.Var("resolution", q.Resolution)
	req.Var("last", q.Last)

	var h HomeConsumptionResponse
	if err := t.APIClient.Run(ctx, req, &h); err != nil {
		log.Fatal(err)
	}
	return h
}

func (p *Price) query(ctx context.Context, t *Client) PriceResponse {
	req := graphql.NewRequest(`
		query ($resolution: PriceInfoResolution) {
			viewer {
				homes {
					currentSubscription{
						priceInfo(resolution: $resolution){
							current{
								total
								energy
								tax
								startsAt
							}
							today {
								total
								energy
								tax
								startsAt
							}
							tomorrow {
								total
								energy
								tax
								startsAt
							}
						}
					}
				}
			}
		}
	`)
	if p.Resolution != "" {
		req.Var("resolution", p.Resolution)
	}
	var price PriceResponse
	if err := t.APIClient.Run(ctx, req, &price); err != nil {
		log.Fatal(err)
	}
	return price
}
