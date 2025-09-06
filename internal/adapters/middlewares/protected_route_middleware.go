package middlewares

import (
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/infrastructure/auth"
	"github.com/CSKU-Lab/main-server/internal/converter"
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
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		if claims, ok := token.Claims.(*auth.JWTClaims); ok && token.Valid {
			roleStringSlice, err := converter.ToStringSlice(claims.Roles)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Something went wrong"})
			}

			roleSlice, err := converter.ToRoleSlice(roleStringSlice)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Something went wrong"})
			}

			c.Locals("user", &models.User{
				ID:           claims.Subject,
				Username:     claims.Username,
				DisplayName:  claims.DisplayName,
				ProfileImage: claims.ProfileImage,
				Roles:        roleSlice,
			})

			return c.Next()
		}

		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
	}
}
