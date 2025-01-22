package utils

import (
	"log"
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
