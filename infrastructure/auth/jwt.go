package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	SignAccessToken(user *models.User) (string, error)
	SignRefreshToken(userID string) (string, error)
}

type JWTClaims struct {
	Username     string  `json:"username"`
	DisplayName  string  `json:"displayName"`
	ProfileImage *string `json:"profileImage"`
	Roles        []any   `json:"roles"`
	jwt.RegisteredClaims
}

func SignAccessToken(user *models.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":          user.ID,
		"username":     user.Username,
		"displayName":  user.DisplayName,
		"profileImage": user.ProfileImage,
		"roles":        user.Roles,
		"iss":          "cs-lab-backend",
		"exp":          time.Now().Add(time.Hour * 1).Unix(),
	})

	return token.SignedString([]byte(secret))

}

func SignRefreshToken(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"iss": "cs-lab-backend",
		"exp": time.Now().Add(time.Hour * 24 * 5).Unix(),
	})

	return token.SignedString([]byte(secret))

}

func VerifyToken(tokenString string, secret string) error {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil
		}
		return err
	}

	return nil
}

func GetClaims(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func IsExpired(token, secret string) bool {
	if token == "" {
		return true
	}

	_, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return true
		}
	}

	return false
}
