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

func blackScholes(spotPrice, strikePrice float64, expirationDate int64, volatility float64, riskFreeRate float64, OptionType string) float64 {
	timeForExpiration := float64(expirationDate) / 365.0

	/*d1 and d2 Calculation*/
	d1 := (math.Log(spotPrice/strikePrice) + (riskFreeRate+0.5*volatility*volatility)*timeForExpiration) / (volatility * math.Sqrt(timeForExpiration))
	d2 := d1 - volatility*math.Sqrt(timeForExpiration)

	/*Option Price Calculation using Black-Scholes formula*/
	if OptionType == "CALL" {
		callPrice := spotPrice*math.Exp(-riskFreeRate*timeForExpiration)*normalCDF(d1) - strikePrice*math.Exp(-riskFreeRate*timeForExpiration)*normalCDF(d2)
		return callPrice
	} else if OptionType == "PUT" {
		putPrice := strikePrice*math.Exp(-riskFreeRate*timeForExpiration)*normalCDF(-d2) - spotPrice*math.Exp(-riskFreeRate*timeForExpiration)*normalCDF(-d1)
		return putPrice
	} else {
		return 0.0
	}
}

func calculateOptionPrice(spot float64, strike float64, timeToExpiry float64, volatility float64, riskFreeRate float64, OptionType string) float64 {
	return blackScholes(spot, strike, int64(timeToExpiry), volatility, riskFreeRate, OptionType)
}

func calculateVolatility(optionPrice float64, timeToExpiry float64, strike float64, spot float64, OptionType string) float64 {
	if timeToExpiry <= 0 {
		return 0.0
	}
	if OptionType != "CALL" && OptionType != "PUT" {
		fmt.Println("Please Provide CALL/PUT Option Type for Black-Scholes Price Calculation")
		return 0.0
	}

	riskFreeRate := 0.05
	maxIterations := 1000
	leastDiff := 10000.0
	volatilityResult := 0.0
	low := 0.0
	high := 1e6 + 7 /*Assuming an initial range for volatility*/

	for i := 0; i < maxIterations; i++ {
		if low >= high {
			break
		}

		mid := low + (high-low)/2.0
		simulatedOptionPrice := calculateOptionPrice(spot, strike, timeToExpiry, mid, riskFreeRate, OptionType)
		/*fmt.Printf("optionPrice for %.2f %.2f %.2f %.2f %.2f %.2f %.2f %.2f: %.2f\n", spot, strike, timeToExpiry, low, mid, high, riskFreeRate, OptionType, simulatedOptionPrice)*/

		if math.Abs(simulatedOptionPrice-optionPrice) < leastDiff {
			leastDiff = math.Abs(simulatedOptionPrice - optionPrice)
			volatilityResult = mid
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
	volatilityResult = low
	/*fmt.Printf("vol %.2f", vol)*/
	return volatilityResult
}

func fetchDeribitIV(symbol, expiry string, strike float64) (float64, error) {
	baseURL := "https://www.deribit.com/api/v2/public/get_historical_volatility"

	/*Create Request URL*/
	params := url.Values{}
	params.Set("currency", symbol)

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	/*Send HTTP request*/
	resp, err := http.Get(url)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	/*Read Response Body*/
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0.0, err
	}

	/*Parse Response JSON*/
	var response DeribitResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0.0, err
	}

	/*Find matching option instrument & get IV*/
	return response.Result[len(response.Result)-1][1], err
}

func compareWithVolatilitySurface(symbol, expiry string, strike, deribitIV float64) (float64, error) {
	if math.Abs(deribitIV-volatilitySurface[symbol][expiry][strike])/deribitIV > 0.02 {
		fmt.Println("More than 2% Deviation")
		return deribitIV, fmt.Errorf("more than 2 percent deviation")
	}
	return volatilitySurface[symbol][expiry][strike], nil
}

// func getVolatilityFromSurface(symbol, expiry string, strike, spot float64, OptionType string) float64 {
// 	/*If Volatility exists in the Surface Check*/
// 	if volSurface, ok := volatilitySurface[symbol]; ok {
// 		if expirySurface, ok := volSurface[expiry]; ok {
// 			if strikeVolatility, ok := expirySurface[strike]; ok {
// 				return strikeVolatility
// 			} else {
// 				expiryDate, err := time.Parse("2006-01-02", expiry)
// 				if err != nil {
// 					print(err)
// 					return 0
// 				}
// 				timeToExpiry := time.Until(expiryDate).Hours() / 24
// 				fmt.Println(timeToExpiry)
// 				volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, OptionType)
// 				fmt.Println(volatilitySurface[symbol][expiry][strike])
// 			}
// 		} else {
// 			volatilitySurface[symbol][expiry] = make(map[float64]float64)
// 			expiryDate, err := time.Parse("2006-01-02", expiry)
// 			if err != nil {
// 				print(err)
// 				return 0
// 			}
// 			timeToExpiry := time.Until(expiryDate).Hours() / 24
// 			fmt.Println(timeToExpiry)
// 			volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, OptionType)
// 			fmt.Println(volatilitySurface[symbol][expiry][strike])
// 		}
// 	} else {
// 		volatilitySurface[symbol] = make(map[string]map[float64]float64)
// 		volatilitySurface[symbol][expiry] = make(map[float64]float64)
// 		expiryDate, err := time.Parse("2006-01-02", expiry)
// 		if err != nil {
// 			print(err)
// 			return 0
// 		}
// 		timeToExpiry := time.Until(expiryDate).Hours() / 24
// 		fmt.Println(timeToExpiry)
// 		volatilitySurface[symbol][expiry][strike] = calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, OptionType)
// 		fmt.Println(volatilitySurface[symbol][expiry][strike])
// 	}

// 	/*Fetch the latest IV data from Deribit's API as a backstop*/
// 	deribitIV, _err := fetchDeribitIV(symbol, expiry, strike)
// 	if _err != nil {
// 		print(_err)
// 		return 0.0
// 	}
// 	/*Compare with existing volatility surface*/
// 	if vol, err := compareWithVolatilitySurface(symbol, expiry, strike, deribitIV); err == nil {
// 		return vol
// 	}
// 	/*If volatility not found in the surface & backstop, return default*/
// 	return volatilitySurface[symbol][expiry][strike]
// }

func getVolatilityFromSurface(symbol, expiry string, strike, spot float64, OptionType string) float64 {
	/*If volatility exists in the surface Check*/
	if volSurface, ok := volatilitySurface[symbol]; ok {
		if expirySurface, ok := volSurface[expiry]; ok {
			if strikeVolatility, ok := expirySurface[strike]; ok {
				return strikeVolatility
			}
		} else {
			volatilitySurface[symbol][expiry] = make(map[float64]float64)
		}
	} else {
		volatilitySurface[symbol] = make(map[string]map[float64]float64)
		volatilitySurface[symbol][expiry] = make(map[float64]float64)
	}

	/*Time to Expiry Calculation*/
	expiryDate, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	timeToExpiry := time.Until(expiryDate).Hours() / 24

	/*Calculate & store volatility in the surface*/
	volatility := calculateVolatility(spot/10, float64(timeToExpiry), strike, spot, OptionType)
	volatilitySurface[symbol][expiry][strike] = volatility

	/*Fetch the latest IV data from Deribit's API as a backstop*/
	deribitIV, err := fetchDeribitIV(symbol, expiry, strike)
	if err != nil {
		fmt.Println(err)
		return 0.0
	}

	/*Compare with existing volatility surface*/
	if vol, err := compareWithVolatilitySurface(symbol, expiry, strike, deribitIV); err == nil {
		return vol
	}

	/*If volatility not found in the surface & backstop, return default*/
	return volatilitySurface[symbol][expiry][strike]
}

func updateVolatilitySurface(symbol, expiry string, strike, spot, lastTrade float64, OptionType string) {
	/*Update the volatility surface with the latest trade*/
	if _, ok := volatilitySurface[symbol]; !ok {
		volatilitySurface[symbol] = make(map[string]map[float64]float64)
	}
	if _, ok := volatilitySurface[symbol][expiry]; !ok {
		volatilitySurface[symbol][expiry] = make(map[float64]float64)
	}
	/*calculate time to expiry*/
	expiryDate := parseDate(expiry)

	timeToExpiry := time.Until(expiryDate).Hours() / 24
	volatilitySurface[symbol][expiry][strike] = calculateVolatility(lastTrade, float64(timeToExpiry), strike, spot, OptionType)
}
