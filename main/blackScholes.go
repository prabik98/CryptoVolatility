package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

var volatilitySurface map[string]map[string]map[float64]float64

func getVolatility(w http.ResponseWriter, r *http.Request) {
	/*Parse Request Payload*/
	var requestPayload GetVolatilityRequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, err.Error()+" -", http.StatusBadRequest)
		return
	}
	optionType := 0.0

	/*Get volatility from the volatility surface*/
	volatility := getVolatilityFromSurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, optionType)

	/*Create Response Payload*/
	response := OptionVolatility{
		Symbol:     requestPayload.Symbol,
		Expiration: parseDate(requestPayload.Expiry),
		Strike:     requestPayload.Strike,
		Volatility: volatility,
		Timestamp:  time.Now(),
	}

	/*Send Response*/
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateVolatility(w http.ResponseWriter, r *http.Request) {
	/*Parse Request Payload*/
	w.Header().Set("Content-Type", "application/json")

	/*Authenticatication*/
	if !authenticate(w, r) {
		return
	}
	var requestPayload UpdateVolatilityRequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	optionType := 0.0

	/*Update Volatility Surface with Latest Trade*/
	updateVolatilitySurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, requestPayload.LastTrade, optionType)

	/*Send Response*/
	json.NewEncoder(w).Encode("Volatility Updation Successful")
	w.WriteHeader(http.StatusOK)
}

func calculateVolatility(optionPrice float64, timeToExpiry float64, strike float64, spot float64, optionType float64) float64 {
	if timeToExpiry <= 0 {
		return 0.0
	}

	riskFreeRate := 0.05

	maxIterations := 1000
	leastDiff := 10000.0
	volatility := 0.0

	low := 0.0
	high := 1e6 + 7 /*Assuming an initial range for volatility*/

	for i := 0; i < maxIterations; i++ {
		if low >= high {
			break
		}

		mid := low + (high-low)/2.0
		simulatedOptionPrice := calculateOptionPrice(spot, strike, timeToExpiry, mid, riskFreeRate, optionType)
		/*fmt.Printf("optionPrice for %.2f %.2f %.2f %.2f %.2f %.2f %.2f %.2f: %.2f\n", spot, strike, timeToExpiry, low, mid, high, riskFreeRate, optionType, simulatedOptionPrice)*/

		if math.Abs(simulatedOptionPrice-optionPrice) < leastDiff {
			leastDiff = math.Abs(simulatedOptionPrice - optionPrice)
			volatility = mid
			/*fmt.Printf("vol %.2f", vol)*/
		}
		if simulatedOptionPrice == optionPrice {
			break
		}
		if simulatedOptionPrice < optionPrice {
			low = mid
			/*fmt.Printf("low %.2f", low)*/
		} else {
			high = mid
			/*fmt.Printf("high %.2f", high)*/
		}
	}
	volatility = low
	/*fmt.Printf("vol %.2f", vol)*/
	return volatility
}

func calculateOptionPrice(spot, strike, timeToExpiry, volatility, riskFreeRate, optionType float64) float64 {

	return blackScholes(spot, strike, int(timeToExpiry), volatility, riskFreeRate, optionType)
}

func blackScholes(spotPrice, strikePrice float64, expirationDate int, volatility, riskFreeRate float64, optionType float64) float64 {
	// Constants
	timeToExpiration := float64(expirationDate) / 365.0

	// Calculate d1 and d2
	d1 := (math.Log(spotPrice/strikePrice) + (riskFreeRate+0.5*volatility*volatility)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))
	d2 := d1 - volatility*math.Sqrt(timeToExpiration)

	// Calculate option price using Black-Scholes formula
	if optionType == 1 {
		callPrice := spotPrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d1) - strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(d2)
		return callPrice
	} else {
		putPrice := strikePrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(-d2) - spotPrice*math.Exp(-riskFreeRate*timeToExpiration)*normalCDF(-d1)
		return putPrice
	}
}

func normalCDF(x float64) float64 {
	// return 0.5 * (1 + math.Erf(x/math.Sqrt2))
	return 0.5 * math.Erfc(-(x)/(math.Sqrt2))
}

func getVolatilityFromSurface(symbol, expiry string, strike, spot float64, optionType float64) float64 {
	// Check if volatility exists in the surface
	if volSurface, ok := volatilitySurface[symbol]; ok {
		if expirySurface, ok := volSurface[expiry]; ok {
			if strikeVolatility, ok := expirySurface[strike]; ok {
				return strikeVolatility
			} else {
				expiryDate, err := time.Parse("2006-01-02", expiry)
				if err != nil {
					print(err)
					return 0
				}
				timeToExpiry := time.Until(expiryDate).Hours() / 24
				fmt.Println(timeToExpiry)
				volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, optionType)
				fmt.Println(volatilitySurface[symbol][expiry][strike])
			}
		} else {
			volatilitySurface[symbol][expiry] = make(map[float64]float64)
			expiryDate, err := time.Parse("2006-01-02", expiry)
			if err != nil {
				print(err)
				return 0
			}
			timeToExpiry := time.Until(expiryDate).Hours() / 24
			fmt.Println(timeToExpiry)
			volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, optionType)
			fmt.Println(volatilitySurface[symbol][expiry][strike])
		}
	} else {
		volatilitySurface[symbol] = make(map[string]map[float64]float64)
		volatilitySurface[symbol][expiry] = make(map[float64]float64)
		expiryDate, err := time.Parse("2006-01-02", expiry)
		if err != nil {
			print(err)
			return 0
		}
		timeToExpiry := time.Until(expiryDate).Hours() / 24
		fmt.Println(timeToExpiry)
		volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, optionType)
		fmt.Println(volatilitySurface[symbol][expiry][strike])
	}

	// Fetch the latest IV data from Deribit's API as a backstop
	deribitIV, _err := fetchDeribitIV(symbol, expiry, strike)
	if _err != nil {
		print(_err)
		return 0.0
	}
	// Compare with existing volatility surface
	if vol, err := compareWithVolatilitySurface(symbol, expiry, strike, deribitIV); err == nil {
		return vol
	}
	// If volatility not found in the surface and backstop, return a default value

	return volatilitySurface[symbol][expiry][strike]
}

type DeribitResponse struct {
	Result [][]float64 `json:"result"`
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
	timeToExpiry := time.Until(expiryDate).Hours() / 24
	volatilitySurface[symbol][expiry][strike] = calculateVolatility(lastTrade, float64(timeToExpiry), strike, spot, optionType)
}

func fetchDeribitIV(symbol, expiry string, strike float64) (float64, error) {
	baseURL := "https://www.deribit.com/api/v2/public/get_historical_volatility"

	// Create the request URL
	params := url.Values{}
	params.Set("currency", symbol)

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Send the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, err
	}

	// Parse the response JSON
	var response DeribitResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0.0, err
	}

	// Find the matching option instrument and extract the IV
	return response.Result[len(response.Result)-1][1], err

}

func compareWithVolatilitySurface(symbol, expiry string, strike, deribitIV float64) (float64, error) {

	if math.Abs(deribitIV-volatilitySurface[symbol][expiry][strike])/deribitIV > 0.02 {
		fmt.Println("more than 2% deviation")
		return deribitIV, fmt.Errorf("value deviating more than 2%%")
	}
	return volatilitySurface[symbol][expiry][strike], nil
}

func parseDate(dateString string) time.Time {
	res, _ := time.Parse("2006-01-02", dateString)
	return res
}
