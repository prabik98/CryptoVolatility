package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

func rateLimiter(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			message := Message{
				Status: "Request Failed",
				Body:   "The API is at capacity, try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		} else {
			next(w, r)
		}
	})
}

func authenticate(w http.ResponseWriter, r *http.Request) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		fmt.Println("Error parsing basic authorization")
		json.NewEncoder(w).Encode("Provide basic authorization")
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	if u != username {
		fmt.Printf("Username provided is incorrect: %s\n", u)
		json.NewEncoder(w).Encode("Username provided is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	if p != password {
		fmt.Printf("Password provided is incorrect: %s\n", u)
		json.NewEncoder(w).Encode("Password provided is incorrect")
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
