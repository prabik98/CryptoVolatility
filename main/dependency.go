package main

import (
	"math"
	"time"

	"golang.org/x/time/rate"
)

var (
	username = "bikashpradhan"
	password = "incorrect"

	clientsCount = rate.Limit(3)
	burstSize    = 5
)

func parseDate(dateString string) time.Time {
	result, _ := time.Parse("2006-01-02", dateString)
	return result
}

func normalCDF(x float64) float64 {
	return 0.5 * math.Erfc(-(x)/(math.Sqrt2))
}
