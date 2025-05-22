package middlewares

import (
	"fmt"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/infrastructure/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func ProtectedRouteMiddleware(secret string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")

		token, err := jwt.ParseWithClaims(accessToken, &auth.JWTClaims{}, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(secret), nil
		})

		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		if claims, ok := token.Claims.(*auth.JWTClaims); ok && token.Valid {
			roles := make([]string, len(claims.Roles))
			for i, v := range claims.Roles {
				roles[i] = v.(string)
			}

			c.Locals("user", &models.User{
				ID:           claims.Subject,
				Username:     claims.Username,
				DisplayName:  claims.DisplayName,
				ProfileImage: claims.ProfileImage,
				Roles:        roles,
			})

			return c.Next()
		}

		return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
	}
}
