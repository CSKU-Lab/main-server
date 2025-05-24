package rest

import (
	"context"
	"log"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/infrastructure/auth"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/middlewares"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/converter"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// TODO: redirect back to frontend and set cookie
func NewAuthRouter(router fiber.Router, appConfig *configs.Config, userService services.UserService, refreshTokenService services.RefreshTokenService) {
	authRouter := router.Group("/auth")

	googleAuth := auth.NewGoogleAuth(appConfig)

	// Google OAuth2
	authRouter.Get("/sign-in/google", func(c *fiber.Ctx) error {
		url, err := googleAuth.GenerateAuthURL()
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error generating auth url")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		return c.Redirect(url)
	})

	authRouter.Get("/sign-in/google/callback", func(c *fiber.Ctx) error {
		state := c.Query("state")
		if !googleAuth.VerifyState(state) {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.BAD_REQUEST, "Invalid State")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		ctx := context.Background()

		code := c.Query("code")

		userInfo, err := googleAuth.GetUserInfo(ctx, code)
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting user info")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		user, err := userService.GetByEmail(c.Context(), userInfo.Email)
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting user")
			}
			return c.Redirect(appConfig.FRONTEND_URL + "/auth/sign-in?error=UNAUTHORIZED")
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			if appConfig.DEV_MODE {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Expires:  time.Now().Add(time.Hour * 1),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Expires:  time.Now().Add(time.Hour * 24 * 5),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		if appConfig.DEV_MODE {
			return c.JSON(fiber.Map{
				"message":       "OK",
				"access_token":  newAccessToken,
				"refresh_token": newRefreshToken,
			})
		}

		return c.Redirect(appConfig.FRONTEND_URL)
	})

	authRouter.Post("/sign-in/credential", middlewares.ValidateMiddleware(&requests.Credential{}), func(c *fiber.Ctx) error {
		credential := c.Locals("request").(*requests.Credential)

		user, err := userService.GetByUsername(c.Context(), credential.Username)
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		password, err := userService.GetPasswordByID(c.Context(), user.ID)
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		err = bcrypt.CompareHashAndPassword([]byte(password), []byte(credential.Password))
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Expires:  time.Now().Add(time.Hour * 1),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Expires:  time.Now().Add(time.Hour * 24 * 5),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		if appConfig.DEV_MODE {
			return c.JSON(fiber.Map{
				"message":       "OK",
				"access_token":  newAccessToken,
				"refresh_token": newRefreshToken,
			})
		}

		return c.Redirect(appConfig.FRONTEND_URL)
	})

	authRouter.Post("/refresh-token", func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")

		err := auth.VerifyToken(accessToken, appConfig.JWTSecret)
		if err != nil {
			log.Println(err)
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		var user *models.User
		if !auth.IsExpired(accessToken, appConfig.JWTSecret) {
			claims, err := auth.GetClaims(accessToken, appConfig.JWTSecret)
			if err != nil {
				return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
			}

			roles, err := converter.ToStringSlice(claims.Roles)
			if err != nil {
				return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
			}

			user = &models.User{
				ID:           claims.Subject,
				Username:     claims.Username,
				DisplayName:  claims.DisplayName,
				ProfileImage: claims.ProfileImage,
				Roles:        roles,
			}

			if claims.ExpiresAt.Sub(time.Now()).Minutes() > 5 {
				return c.JSON(fiber.Map{
					"message": "token is stil valid",
				})
			}
		} else {
			claims, err := auth.GetClaims(refreshToken, appConfig.JWTRefreshSecret)
			if err != nil {
				return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
			}

			user, err = userService.GetByID(c.Context(), claims.Subject)
			if err != nil {
				return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
			}
		}

		dbRefreshToken, err := refreshTokenService.Get(c.Context(), user.ID)
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		// check for replay attack
		if dbRefreshToken != refreshToken {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Something went wrong")
		}

		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    newAccessToken,
			Expires:  time.Now().Add(time.Hour * 1),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		c.Cookie(&fiber.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Expires:  time.Now().Add(time.Hour * 24 * 5),
			HTTPOnly: true,
			SameSite: fiber.CookieSameSiteLaxMode,
			Domain:   appConfig.COOKIE_DOMAIN,
			Secure:   false,
		})

		if appConfig.DEV_MODE {
			return c.JSON(fiber.Map{
				"message":       "OK",
				"access_token":  newAccessToken,
				"refresh_token": newRefreshToken,
			})
		}

		return c.JSON(fiber.Map{
			"message": "success",
		})
	})

}
