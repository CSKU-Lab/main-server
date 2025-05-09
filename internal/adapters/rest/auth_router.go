package rest

import (
	"context"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/infrastructure/auth"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/middlewares"
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
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error generating auth url")
		}

		return c.Redirect(url)
	})

	authRouter.Get("/sign-in/google/callback", func(c *fiber.Ctx) error {
		state := c.Query("state")
		if !googleAuth.VerifyState(state) {
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		ctx := context.Background()

		code := c.Query("code")

		userInfo, err := googleAuth.GetUserInfo(ctx, code)
		if err != nil {
			return cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Error getting user info")
		}

		user, err := userService.GetByEmail(c.Context(), userInfo.Email)
		if err != nil {
			return c.Redirect(appConfig.FRONTEND_URL + "/auth/sign-in?error=UNAUTHORIZED")
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

	authRouter.Post("/refresh-token", middlewares.ProtectedRouteMiddleware(appConfig.JWTSecret), func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")
		user := c.Locals("user").(*models.User)

		// verify refresh token if valid and not expired
		err := auth.VerifyToken(refreshToken, appConfig.JWTRefreshSecret)
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		claims, err := auth.GetClaims(accessToken, appConfig.JWTSecret)
		if err != nil {
			return cserrors.New(cserrors.UNAUTHORIZED, "Unauthorized")
		}

		if claims.ExpiresAt.Sub(time.Now()).Minutes() > 5 {
			return c.JSON(fiber.Map{
				"message": "token is stil valid",
			})
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
