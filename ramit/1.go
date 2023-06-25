package ramit

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

type OptionVolatility struct {
	Symbol     string    `json:"symbol"`
	Expiration time.Time `json:"expiration"`
	Strike     float64   `json:"strike"`
	Volatility float64   `json:"volatility"`
	Timestamp  int64     `json:"timestamp"`
}

type RequestPayload struct {
	Symbol    string  `json:"symbol"`
	Expiry    string  `json:"expiry"`
	Strike    float64 `json:"strike"`
	Spot      float64 `json:"spot"`
	LastTrade float64 `json:"last_trade"`
}

var volatilitySurface map[string]map[string]map[float64]float64

func main() {
	volatilitySurface = make(map[string]map[string]map[float64]float64)
	// print(calculateVolatility(204, "2023-06-26", 8200, 8400))
	// print(calculateOptionPrice(8400, 8200, 2, 0.18, 0.05, 1))
	fmt.Printf("volatility: %.2f%%\n", calculateOptionPrice(8400, 8100, 2, 0.18, 0.05, 1))
	fmt.Printf("volatility: %.2f%%\n", calculateVolatility(303, 2, 8100, 8400, 1))

	// http.HandleFunc("/volatility", getVolatility)
	// http.HandleFunc("/update", updateVolatility)
	// log.Fatal(http.ListenAndServe(":8080", nil))
}

func getVolatility(w http.ResponseWriter, r *http.Request) {
	// Parse request payload
	var requestPayload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	optionType := 1.0

	// Get volatility from the volatility surface
	volatility := getVolatilityFromSurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, optionType)

	// Create response payload
	response := OptionVolatility{
		Symbol:     requestPayload.Symbol,
		Expiration: parseDate(requestPayload.Expiry),
		Strike:     requestPayload.Strike,
		Volatility: volatility,
		Timestamp:  time.Now().Unix(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateVolatility(w http.ResponseWriter, r *http.Request) {
	// Parse request payload
	var requestPayload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	optionType := 1.0

	// Update the volatility surface with the latest trade
	updateVolatilitySurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, requestPayload.LastTrade, optionType)

	// Send response
	w.WriteHeader(http.StatusOK)
}

func calculateVolatility(optionPrice float64, timeToExpiry float64, strike float64, spot float64, optionType float64) float64 {
	// Convert expiry date string to time.Time
	// expiryDate, err := time.Parse("2006-01-02", expiry)
	// if err != nil {
	// 	return 0.0
	// }
	// riskFreeRate := 0.07

	// // Calculate time to expiry in years
	// timeToExpiry := float64(expiryDate.Sub(time.Now())) / float64(time.Hour*24*365)

	// // Initialize implied volatility
	// impliedVolatility := 0.5
	// epsilon := 0.0001

	// // Calculate option price using initial implied volatility
	// initialOptionPrice := calculateOptionPrice(spot, strike, timeToExpiry, impliedVolatility, riskFreeRate, 1)

	// // Iteratively adjust implied volatility using the Newton-Raphson method
	// for i := 0; i < maxIterations; i++ {
	// 	v1 := calculateOptionPrice(spot, strike, timeToExpiry, impliedVolatility, riskFreeRate, 1)
	// 	v2 := calculateOptionPrice(spot, strike, timeToExpiry, impliedVolatility, riskFreeRate, -1)

	// 	vega := (v1 - v2) / 2

	// 	// Update implied volatility
	// 	impliedVolatility -= (initialOptionPrice - optionPrice) / vega

	// 	// Recalculate option price with updated implied volatility
	// 	optionPrice = calculateOptionPrice(spot, strike, timeToExpiry, impliedVolatility, riskFreeRate, 1)

	// 	// Check convergence
	// 	if math.Abs(optionPrice-initialOptionPrice) < epsilon {
	// 		break
	// 	}
	// }

	// return impliedVolatility
	riskFreeRate := 0.07

	maxIterations := 1000
	leastDiff := 10000.0
	vol := 0.0

	for i := 0; i < maxIterations; i++ {
		xx := calculateOptionPrice(spot, strike, timeToExpiry, float64(i)/100, riskFreeRate, 1)
		fmt.Printf("optionPrice for %.2f %.2f %.2f %.2f %.2f %.2f: %.2f\n", spot, strike, timeToExpiry, float64(i)/100, riskFreeRate, 1.0, xx)

		if math.Abs(xx-optionPrice) < leastDiff {
			leastDiff = math.Abs(xx - optionPrice)
			vol = float64(i) / 100
		}
	}

	return vol
}

func calculateOptionPrice(spot, strike, timeToExpiry, volatility, riskFreeRate, optionType float64) float64 {
	// d1 := (math.Log(spot/strike) + (riskFreeRate+0.5*math.Pow(volatility, 2))*timeToExpiry) / (volatility * math.Sqrt(timeToExpiry))
	// d2 := d1 - volatility*math.Sqrt(timeToExpiry)

	if optionType == 1 {
		// return spot*normalCDF(d1) - strike*math.Exp(-riskFreeRate*timeToExpiry)*normalCDF(d2)
		return blackScholes(spot, strike, int(timeToExpiry), volatility, riskFreeRate)
	} else {
		return 0.0
	}
}

func blackScholes(spotPrice, strikePrice float64, expirationDate int, volatility float64, riskFreeRate float64) float64 {
	// Constants
	timeToExpiration := float64(expirationDate) / 365.0

	// Calculate d1 and d2
	d1 := (math.Log(spotPrice/strikePrice) + (riskFreeRate+0.5*volatility*volatility)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))
	d2 := d1 - volatility*math.Sqrt(timeToExpiration)

	// Calculate option price using Black-Scholes formula
	callPrice := spotPrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d1) - strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d2)

	return callPrice
}

func normalCDF(x float64) float64 {
	// return 0.5 * (1 + math.Erf(x/math.Sqrt2))
	return 0.5 * math.Erfc(-(x)/(math.Sqrt2))
}

func normalPDF(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}

// func calD1(spotPrice float64, strikePrice float64, volatility float64, expirationDate float64) float64 {
// 	riskFreeRate := 0.05
// 	timeToExpiration := expirationDate / 365.0
// 	return (math.Log(spotPrice/strikePrice) + (riskFreeRate+volatility*volatility/2.0)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))
// }

// func calD2(d1 float64, volatility float64, expirationDate float64) float64 {
// 	timeToExpiration := expirationDate / 365.0
// 	return d1 - (volatility * math.Sqrt(timeToExpiration))
// }

func getVolatilityFromSurface(symbol, expiry string, strike, spot float64, optionType float64) float64 {
	// Check if volatility exists in the surface
	if volSurface, ok := volatilitySurface[symbol]; ok {
		if expirySurface, ok := volSurface[expiry]; ok {
			if strikeVolatility, ok := expirySurface[strike]; ok {
				return strikeVolatility
			}
		}
	}

	// Fetch the latest IV data from Deribit's API as a backstop
	// deribitIV := fetchDeribitIV(symbol, expiry, strike)
	// // Compare with existing volatility surface
	// if vol, err := compareWithVolatilitySurface(symbol, expiry, strike, deribitIV); err == nil {
	// 	return vol
	// }

	// If volatility not found in the surface and backstop, return a default value
	// TODO: calculate time to expiry

	expiryDate, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		print(err)
		return 0
	}
	timeToExpiry := time.Since(expiryDate)
	return calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, optionType)
}

func updateVolatilitySurface(symbol, expiry string, strike, spot, lastTrade float64, optionType float64) {
	// Update the volatility surface with the latest trade
	if _, ok := volatilitySurface[symbol]; !ok {
		volatilitySurface[symbol] = make(map[string]map[float64]float64)
	}
	if _, ok := volatilitySurface[symbol][expiry]; !ok {
		volatilitySurface[symbol][expiry] = make(map[float64]float64)
	}
	// TODO: calculate time to expiry
	expiryDate, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		print(err)
		return
	}
	timeToExpiry := time.Since(expiryDate)
	volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, optionType)
}

func fetchDeribitIV(symbol, expiry string, strike float64) float64 {
	// Fetch the latest implied volatility (IV) data from Deribit's API
	// ...
	// Return the fetched IV
	return 0.0
}

func compareWithVolatilitySurface(symbol, expiry string, strike, deribitIV float64) (float64, error) {
	// Compare the fetched IV with the existing volatility surface
	// If the difference exceeds 2%, update the volatility surface with the new IV
	// ...
	// Return the volatility from the surface or an error if the difference exceeds the threshold
	return 0.0, nil
}

func parseDate(dateString string) time.Time {
	// Parse the date string into a time.Time object
	// ...
	// Return the parsed time.Time object
	return time.Now()
}
