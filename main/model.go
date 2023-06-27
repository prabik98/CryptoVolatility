package main

import "time"

type DeribitResponse struct {
	Result [][]float64 `json:"result"`
}

type OptionVolatility struct {
	Symbol     string    `json:"symbol"`
	Expiration time.Time `json:"expiration"`
	Strike     float64   `json:"strike"`
	Volatility float64   `json:"volatility"`
	Timestamp  time.Time `json:"timestamp"`
}

type UpdateVolatilityRequestPayload struct {
	Symbol     string  `json:"symbol"`
	Expiry     string  `json:"expiry"`
	Strike     float64 `json:"strike,string"`
	Spot       float64 `json:"spot,string"`
	LastTrade  float64 `json:"last_trade,string"`
	OptionType string  `json:"option_type"`
}

type GetVolatilityRequestPayload struct {
	Symbol     string  `json:"symbol"`
	Expiry     string  `json:"expiry"`
	Strike     float64 `json:"strike,string"`
	Spot       float64 `json:"spot,string"`
	OptionType string  `json:"option_type,string"`
}

type Message struct {
	Status string `json:"status"`
	Body   string `json:"body"`
}
