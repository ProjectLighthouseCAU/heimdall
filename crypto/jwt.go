package crypto

import (
	"encoding/hex"
	"fmt"

	"lighthouse.uni-kiel.de/lighthouse-api/config"
)

var JwtPrivateKey = []byte(config.GetString("JWT_PRIVATE_KEY", string(NewRandomBytes(32))))

func init() {
	fmt.Println("[crypto] JWT-Private-Key: ", hex.EncodeToString(JwtPrivateKey))
}
