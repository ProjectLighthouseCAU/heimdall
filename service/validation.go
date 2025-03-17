package service

import (
	"strings"

	"github.com/ProjectLighthouseCAU/heimdall/config"
	"github.com/ProjectLighthouseCAU/heimdall/crypto"
	"github.com/asaskevich/govalidator"
)

func isValidName(str string) bool {
	return govalidator.Matches(str, `^[A-Za-z0-9_@.#&+-]+$`)
}

func isValidEmail(str string) bool {
	return str == "" || govalidator.IsEmail(str)
}

func isValidPassword(str string) bool {
	if strings.TrimSpace(str) == "" {
		return false
	}
	if len(str) < config.MinPasswordLength {
		return false
	}
	if len(str) > crypto.MaxPasswordLength {
		return false
	}
	// TODO: more password criteria
	return true
}

func isValidRegistrationKey(str string) bool {
	return isValidPassword(str) // TODO: separate criteria
}
