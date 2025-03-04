package crypto

import (
	"log"
	"time"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	MaxPasswordLength = 72 // maximum password length for bcrypt (see bcrypt.ErrPasswordTooLong)
	MinBCryptCost     = 12 // recommended by IETF best practices https://www.ietf.org/archive/id/draft-ietf-kitten-password-storage-07.html#name-bcrypt
)

var optimalCost = setCostFactor()

func setCostFactor() int {
	if config.HashBCryptCostFactor < 12 || config.HashBCryptCostFactor > bcrypt.MaxCost {
		return calculateOptimalBCryptCost()
	}
	return config.HashBCryptCostFactor
}

func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), optimalCost)
}

func PasswordMatchesHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Estimates the bcrypt hashing cost factor using a microbenchmark
// Tries to set the cost such that hashing takes ~250ms on the current machine
func calculateOptimalBCryptCost() int {
	log.Println("BCrypt")
	log.Println("	Executing microbenchmark for optimal hashing cost factor...")
	cost := 5
	start := time.Now()
	bcrypt.GenerateFromPassword([]byte("microbenchmark"), cost)
	duration := time.Since(start)
	for duration.Milliseconds() < int64(config.HashingTimeMs) {
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
