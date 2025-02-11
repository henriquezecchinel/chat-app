package utils

import (
	"log"
	"net/http"
	"strconv"
)

// Atoi is a helper function to convert string to integer.
func Atoi(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Println("Failed to convert string to integer:", err)
		return 0, err
	}
	return i, nil
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
