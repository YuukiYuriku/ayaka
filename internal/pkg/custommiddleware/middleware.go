// internal/api/middleware.go
package custommiddleware

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/domain/tbluser"
	"gitlab.com/ayaka/internal/pkg/customerrors"
	jwtconfig "gitlab.com/ayaka/internal/pkg/jwt"
)

type MiddlewareHandler struct {
	Config            *config.Config               `inject:"config"`
	TemplateCacheRepo tbluser.RepositoryLoginCache `inject:"tblUserCacheRepository"`
	Log               *LogActivityHandler          `inject:"logActivity"`
}

func (m *MiddlewareHandler) AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			go m.Log.LogUserInfo("User", "WARN", "Unauthorized user logout: Header is required")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check if the header starts with "Bearer "
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			go m.Log.LogUserInfo("User", "WARN", "Unauthorized user logout: Invalid authorization format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format",
			})
		}

		tokenId := parts[1]

		// Parse dan validasi token
		claims := &jwtconfig.Claims{}
		tokenString, err := m.TemplateCacheRepo.GetToken(c.Context(), tokenId)
		if err != nil {
			go m.Log.LogUserInfo(claims.UserCode, "ERROR", fmt.Sprintf("Could not logout user: %s", err.Error()))
			log.Printf("Detailed error: %+v", err)
		}
		if errors.Is(err, customerrors.ErrDataNotFound) {
			go m.Log.LogUserInfo(claims.UserCode, "WARN", "Token invalid or not found")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validasi signing method secara spesifik
			if token.Method != jwt.SigningMethodHS256 {
				go m.Log.LogUserInfo(claims.UserCode, "WARN", "Unexpected signing method")
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.Config.JWT.JWTKEY), nil
		})

		if err != nil {
			log.Printf("Detailed error: %+v", err)
			go m.Log.LogUserInfo(claims.UserCode, "WARN", "Token invalid or not found")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		if claims, ok := token.Claims.(*jwtconfig.Claims); ok && token.Valid {
			// Store user information in context for later use
			c.Locals("user", claims)
			return c.Next()
		}

		go m.Log.LogUserInfo(claims.UserCode, "WARN", "Token invalid or not found")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}
}
