package config

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// PostgreSQL Database
	DatabaseHost     string = getString("DB_HOST", "localhost")
	DatabasePort     int    = getInt("DB_PORT", 5432)
	DatabaseUser     string = getString("DB_USER", "postgres")
	DatabasePassword string = getString("DB_PASS", "postgres")
	DatabaseName     string = getString("DB_NAME", "heimdall")

	// Redis Database
	RedisHost     string = getString("REDIS_HOST", "127.0.0.1")
	RedisPort     int    = getInt("REDIS_PORT", 6379)
	RedisUser     string = getString("REDIS_USER", "")
	RedisPassword string = getString("REDIS_PASSWORD", "")

	// Crypto hashing
	HashBCryptCostFactor int = getInt("BCRYPT_COST_FACTOR", 0) // if <12 or >31 the cost factor will be calculated based on HASHING_TIME_MS
	HashingTimeMs        int = getInt("HASHING_TIME_MS", 250)  // approximate time for hashing a password (used to calculate the bcrypt cost factor if not provided)

	// API host and path
	ApiHost     string = getString("API_HOST", "https://lighthouse.uni-kiel.de") // for CORS and Swagger UI API documentation
	ApiBasePath string = getString("API_BASE_PATH", "/api")                      // used only for Swagger UI

	ProxyHeader  string = getString("PROXY_HEADER", "X-Real-Ip") // "X-Real-Ip" behind a reverse proxy and "" for hosting without a proxy
	TrustedProxy string = getString("TRUSTED_PROXY", "traefik")
	// Cross-Origin-Resource-Sharing
	CorsAllowOrigins     string = getString("CORS_ALLOW_ORIGINS", ApiHost) // by default only allow the API host, add allowed origins by appending them separated with commas
	CorsAllowCredentials bool   = getBool("CORS_ALLOW_CREDENTIALS", false)

	// Rate limiter
	DisableRateLimiter bool = getBool("DISABLE_RATE_LIMITER", false)

	// Domain specific config
	AdminRoleName          string        = getString("ADMIN_ROLENAME", "admin")
	DeployRoleName         string        = getString("DEPLOY_ROLENAME", "deploy")
	RegistrationKeyLength  int           = getInt("REGISTRATION_KEY_LENGTH", 20)
	ApiTokenExpirationTime time.Duration = getDuration("API_TOKEN_EXPIRATION_TIME", 3*24*time.Hour)
	MinPasswordLength      int           = getInt("MIN_PASSWORD_LENGTH", 12)
	InternalIPs            []net.IP      = parseIPs(getString("INTERNAL_IPS", ""))

	UseTestDatabase bool = getBool("USE_TEST_DATABASE", false) // TODO: remove in prod - this function deletes the whole database
)

func getString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		s, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Found Config %s=%s, but could not parse it (int required)", key, value)
			return defaultValue
		}
		return s
	}
	return defaultValue
}

func getBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		s, err := strconv.ParseBool(value)
		if err != nil {
			log.Printf("Found Config %s=%s, but could not parse it (bool required)", key, value)
			return defaultValue
		}
		return s
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		d, err := time.ParseDuration(value)
		if err != nil {
			log.Printf("Found Config %s=%s, but could not parse it (duration required, e.g. \"1s\")", key, value)
			return defaultValue
		}
		return d
	}
	return defaultValue
}

func parseIPs(ipsString string) []net.IP {
	var ips []net.IP
	if ipsString == "" {
		return ips
	}
	ipStrings := strings.Split(ipsString, ",")
	for _, ipString := range ipStrings {
		ip := net.ParseIP(ipString)
		if ip == nil {
			panic("Error: could not parse IP from INTERNAL_IPS: " + ipString)
		}
		ips = append(ips, ip)
	}
	return ips
}
