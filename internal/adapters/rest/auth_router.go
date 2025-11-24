package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/infrastructure/auth"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/converter"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func NewAuthRouter(router fiber.Router, appConfig *configs.Config, userService services.UserService, refreshTokenService services.RefreshTokenService) {
	authRouter := router.Group("/auth")

	googleAuth := auth.NewGoogleAuth(appConfig)

	// Google OAuth2
	authRouter.Get("/sign-in/google", func(c *fiber.Ctx) error {
		url, err := googleAuth.GenerateAuthURL()
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error generating auth url"})
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		return c.Redirect(url)
	})

	authRouter.Get("/sign-in/google/callback", func(c *fiber.Ctx) error {
		state := c.Query("state")
		if !googleAuth.VerifyState(state) {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid State"})
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		ctx := context.Background()

		code := c.Query("code")

		userInfo, err := googleAuth.GetUserInfo(ctx, code)
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting user info"})
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		user, err := userService.GetByEmail(c.Context(), userInfo.Email)
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error getting user"})
			}
			return c.Redirect(appConfig.FRONTEND_URL + "/auth/sign-in?error=UNAUTHORIZED")
		}

		if user.ProfileImage == nil {
			err = userService.Update(c.Context(), user.ID, &requests.UpdateUser{
				ProfileImage: &userInfo.ProfileImage,
			})
			if err != nil {
				if appConfig.DevMode {
					return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Error updating user profile image"})
				}
				return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
			}
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
			}
			return cserrors.NewRedirect(cserrors.REDIRECT_SOMETHING_WENT_WRONG)
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			if appConfig.DevMode {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
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

		if appConfig.DevMode {
			return c.JSON(fiber.Map{
				"message":       "OK",
				"access_token":  newAccessToken,
				"refresh_token": newRefreshToken,
			})
		}

		return c.Redirect(appConfig.FRONTEND_URL)
	})

	authRouter.Post("/sign-in/credential", middlewares.ValidateMiddleware[requests.Credential](), func(c *fiber.Ctx) error {
		credential := c.Locals("body").(*requests.Credential)

		user, err := userService.GetByUsername(c.Context(), credential.Username)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		password, err := userService.GetPasswordByID(c.Context(), user.ID)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		err = bcrypt.CompareHashAndPassword([]byte(password), []byte(credential.Password))
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
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

		return c.JSON(fiber.Map{
			"message": "success",
		})
	})

	authRouter.Post("/refresh-token", func(c *fiber.Ctx) error {
		accessToken := c.Cookies("access_token")
		refreshToken := c.Cookies("refresh_token")

		var user *models.User
		if !auth.IsExpired(accessToken, appConfig.JWTSecret) {
			err := auth.VerifyToken(accessToken, appConfig.JWTSecret)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
			}

			claims, err := auth.GetClaims(accessToken, appConfig.JWTSecret)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
			}

			roleStringSlice, err := converter.ToStringSlice(claims.Roles)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
			}

			roles, err := converter.ToRoleSlice(roleStringSlice)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
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
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
			}

			user, err = userService.GetByID(c.Context(), claims.Subject)
			if err != nil {
				return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
			}
		}

		dbRefreshToken, err := refreshTokenService.Get(c.Context(), user.ID)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		// check for replay attack
		if dbRefreshToken != refreshToken {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusUnauthorized, Message: "Unauthorized"})
		}

		newAccessToken, err := auth.SignAccessToken(user, appConfig.JWTSecret)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
		}

		newRefreshToken, err := auth.SignRefreshToken(user.ID, appConfig.JWTRefreshSecret)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
		}

		err = refreshTokenService.Set(c.Context(), user.ID, newRefreshToken)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Something went wrong"})
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

		if appConfig.DevMode {
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
