package middleware

import (
	jwtware "github.com/gofiber/jwt/v3"
	"lighthouse.uni-kiel.de/lighthouse-api/crypto"
)

var JwtMiddleware = jwtware.New(jwtware.Config{
	SigningKey: crypto.JwtPrivateKey,
})
