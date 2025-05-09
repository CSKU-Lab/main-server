package main

import (
	"fmt"
	"log"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/middlewares"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/rest"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/sqlx"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	userRepo := sqlx.NewSqlxUserRepository(db)
	userService := services.NewUserService(userRepo)

	refreshTokenRepo := sqlx.NewSQLxRefreshTokenRepository(db)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	semesterRepo := sqlx.NewSqlxSemesterRepository(db)
	semesterService := services.NewSemesterService(semesterRepo)

	courseRepo := sqlx.NewSqlxCourseRepository(db)
	courseService := services.NewCourseService(courseRepo)

	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
	})

	api := app.Group("/api/v1")

	rest.NewAuthRouter(api, config, userService, refreshTokenService)

	protectedApi := api.Group("/", middlewares.ProtectedRouteMiddleware(config.JWTSecret))

	rest.NewAdminRouter(&rest.AdminRouter{
		Router:          protectedApi,
		UserService:     userService,
		SemesterService: semesterService,
		CourseService:   courseService,
	})

	port := fmt.Sprintf(":%v", config.Port)

	err := app.Listen(port)
	if err != nil {
		log.Fatal("Error starting server on Port ", port)
	}

}
