package main

import (
	"fmt"
	"math"
)

const (
	epsilon       = 1e-6 // Tolerance for numerical convergence
	maxIterations = 10   // Maximum iterations for convergence
	initialVol    = 0.17 // Initial guess for volatility
	riskFreeRate  = 0.07 // Risk-free interest rate
)

func main() {
	// Input parameters
	spotPrice := 8400.0
	strikePrice := 8000.0
	callOptionPrice := 492.31
	//putOptionPrice := 7.0
	expirationDate := 2.0 // Time to expiration in days

	// Calculate implied volatility for call option
	callVolatility := calculateVolatility(spotPrice, strikePrice, callOptionPrice, expirationDate, "call")
	fmt.Printf("Call Option - Implied Volatility: %.2f%%\n", callVolatility*100)

}

func calculateVolatility(spotPrice, strikePrice, optionPrice, expirationDate float64, optionType string) float64 {

	volatility := initialVol

	for i := 0; i < maxIterations; i++ {
		// Calculate option price with current volatility
		calculatedOptionPrice := 0.0

		// if optionType == "call" {
		calculatedOptionPrice = blackScholes(spotPrice, strikePrice, expirationDate, volatility)

		vega := calculateVega(spotPrice, strikePrice, volatility, expirationDate) * 0.01
		fmt.Printf("vega: %.2f%%\n", vega*100)
		fmt.Printf("calculatedOptionPrice: %.2f%%\n", calculatedOptionPrice)
		// Update volatility using Newton-Raphson method
		if vega != 0 {
			volatility = volatility - (calculatedOptionPrice-optionPrice)/vega
		}
		// Check for convergence
		if math.Abs(calculatedOptionPrice-optionPrice) < epsilon {
			break
		}
		fmt.Printf("volatility : %.2f%%\n", volatility*100)
	}

	return volatility
}

func blackScholes(spotPrice, strikePrice float64, expirationDate float64, volatility float64) float64 {
	// Constants
	riskFreeRate := 0.07
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

func calculateVega(spotPrice, strikePrice float64, volatility float64, expirationDate float64) float64 {
	// Constants
	timeToExpiration := float64(expirationDate) / 365.0

	// Calculate d1
	d1 := (math.Log(spotPrice/strikePrice) + (riskFreeRate+0.5*volatility*volatility)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))

	// Calculate vega using Black-Scholes formula
	vega := spotPrice * math.Sqrt(timeToExpiration) * normalPDF(d1) * 0.01
	// fmt.Printf("d1: %.2f%%\n", d1)
	// fmt.Printf("normalPDF: %.2f%%\n", normalPDF(d1))
	return vega
}

func calD1(spotPrice float64, strikePrice float64, volatility float64, expirationDate float64) float64 {
	timeToExpiration := expirationDate / 365.0
	return (math.Log(spotPrice/strikePrice) + (riskFreeRate+volatility*volatility/2.0)*timeToExpiration) / (volatility * math.Sqrt(timeToExpiration))
}

func calD2(d1 float64, volatility float64, expirationDate float64) float64 {
	timeToExpiration := expirationDate / 365.0
	return d1 - (volatility * math.Sqrt(timeToExpiration))
}
