package auth 

import (
	"os"
	"strings"

    "github.com/gofiber/fiber/v2"
)

func checkUser(pin string) (string, bool) {
	userPins := os.Getenv("USER_PINS")
	pairs := strings.Split(userPins, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			username := strings.TrimSpace(parts[0])
			validPin := strings.TrimSpace(parts[1])
			if pin == validPin {
				return username, true
			}
		}
	}
	return "", false
}

// Login endpoint
func LoginHandler(c *fiber.Ctx) error {
	pin := c.Query("pin")
	if pin == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "PIN required"})
	}

	user, ok := checkUser(pin)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Invalid PIN"})
	}

	return c.JSON(fiber.Map{"success": true, "user": user})
}

// Middleware to protect routes
func PinAuthMiddleware(c *fiber.Ctx) error {
	pin := c.Query("pin")
	user, ok := checkUser(pin)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid PIN"})
	}
	c.Locals("user", user)
	return c.Next()
}
