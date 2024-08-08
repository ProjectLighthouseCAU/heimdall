package crypto

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
)

var optimalCost = calculateOptimalCost()

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), optimalCost)
}

func PasswordMatchesHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Estimates the bcrypt hashing cost factor using a microbenchmark
// Tries to set the cost such that hashing takes ~250ms on the current machine
func calculateOptimalCost() int {
	log.Println("BCrypt")
	log.Println("	Executing microbenchmark for optimal hashing cost factor...")
	cost := 5
	start := time.Now()
	bcrypt.GenerateFromPassword([]byte("microbenchmark"), cost)
	duration := time.Since(start)
	for duration.Milliseconds() < int64(config.GetInt("HASHING_TIME_MS", 250)) {
		cost += 1
		duration *= 2
	}
	// the minimum hashing cost should be 10 for security reasons
	if cost < 10 {
		cost = 10
	}
	log.Printf("	Setting optimal bcrypt hashing cost factor to: %d\n", cost)
	return cost
}
