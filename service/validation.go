package service

import (
	"log"
	"strings"

	"github.com/asaskevich/govalidator"
	"lighthouse.uni-kiel.de/lighthouse-api/config"
)

func isValidName(str string) bool {
	return govalidator.Matches(str, `^[A-Za-z0-9_@.#&+-]+$`)
}

func isValidEmail(str string) bool {
	return govalidator.IsEmail(str)
}

func isValidPassword(str string) bool {
	if strings.TrimSpace(str) == "" {
		log.Println("pw empty")
		return false
	}
	if len(str) < config.GetInt("MIN_PASSWORD_LENGTH", 12) {
		log.Println("pw too short")
		return false
	}
	// TODO: more password criteria
	return true
}

func isValidRegistrationKey(str string) bool {
	return isValidPassword(str) // TODO: separate criteria
}
