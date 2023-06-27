package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func getVolatility(w http.ResponseWriter, r *http.Request) {
	/*Parse Request Payload*/
	var requestPayload GetVolatilityRequestPayload
	err := json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, err.Error()+" -", http.StatusBadRequest)
		return
	}

	/*Get volatility from the volatility surface*/
	volatility := getVolatilityFromSurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, requestPayload.OptionType)

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

	/*Update Volatility Surface with Latest Trade*/
	updateVolatilitySurface(requestPayload.Symbol, requestPayload.Expiry, requestPayload.Strike, requestPayload.Spot, requestPayload.LastTrade, requestPayload.OptionType)

	/*Send Response*/
	json.NewEncoder(w).Encode("Volatility Updation Successful")
	w.WriteHeader(http.StatusOK)
}

func main() {
	db := connectToDatabase()
	defer db.Close()
	volatilitySurface = make(map[string]map[string]map[float64]float64)
	fmt.Printf("volatility: %.2f%%\n", calculateVolatility(303, 2, 8100, 8400, "CALL"))

	http.Handle("/volatility", rateLimiter(getVolatility))
	http.Handle("/update", rateLimiter(updateVolatility))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
