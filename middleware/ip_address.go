package middleware

import (
	"net"
	"slices"

	"github.com/gofiber/fiber/v2"
)

// Make sure that c.IP() returns the correct client IP
// Set PROXY_HEADER correctly "X-Real-Ip" behind a reverse proxy and "" for hosting without a proxy

func AllowLoopbackAndPrivateIPs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIp := net.ParseIP(c.IP())
		if clientIp.IsPrivate() || clientIp.IsLoopback() {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusForbidden)
	}
}

func AllowIPs(ips []net.IP) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIp := net.ParseIP(c.IP())
		if slices.ContainsFunc(ips, func(ip net.IP) bool {
			return slices.Equal(ip, clientIp)
		}) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusForbidden)
	}
}

func AllowLoopbackAndPrivateIPsAnd(ips []net.IP) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIp := net.ParseIP(c.IP())
		if clientIp.IsPrivate() || clientIp.IsLoopback() {
			return c.Next()
		}
		if slices.ContainsFunc(ips, func(ip net.IP) bool {
			return slices.Equal(ip, clientIp)
		}) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusForbidden)
	}
}
