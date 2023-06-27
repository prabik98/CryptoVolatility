package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	db := connectToDatabase()
	defer db.Close()
	volatilitySurface = make(map[string]map[string]map[float64]float64)
	fmt.Printf("volatility: %.2f%%\n", calculateVolatility(303, 2, 8100, 8400, 1))

	http.Handle("/volatility", rateLimiter(getVolatility))
	http.Handle("/update", rateLimiter(updateVolatility))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
