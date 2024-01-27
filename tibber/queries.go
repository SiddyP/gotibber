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

type Price struct {
}

type WebsocketSubscriptionUrl struct {
}

func setHeaders(r *graphql.Request, t *Client) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer "+t.APIClient.Config.Token)
}

func (w *WebsocketSubscriptionUrl) query(ctx context.Context, t *Client) WebsocketSubscriptionUrlResponse {
	req := graphql.NewRequest(`
		query {
			viewer {
				websocketSubscriptionUrl
			}
		}
		`)
	setHeaders(req, t)

	var ws WebsocketSubscriptionUrlResponse
	if err := t.APIClient.GQLClient.Run(ctx, req, &ws); err != nil {
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
	setHeaders(req, t)
	// run it and capture the response
	var u UserResponse
	if err := t.APIClient.GQLClient.Run(ctx, req, &u); err != nil {
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
	setHeaders(req, t)

	// set any variables
	req.Var("resolution", q.Resolution)
	req.Var("last", q.Last)

	var h HomeConsumptionResponse
	if err := t.APIClient.GQLClient.Run(ctx, req, &h); err != nil {
		log.Fatal(err)
	}
	return h
}

func (p *Price) query(ctx context.Context, t *Client) PriceResponse {
	req := graphql.NewRequest(`
		query {
				viewer {
				homes {
					currentSubscription{
					priceInfo{
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
	setHeaders(req, t)

	var price PriceResponse
	if err := t.APIClient.GQLClient.Run(ctx, req, &price); err != nil {
		log.Fatal(err)
	}
	return price
}
