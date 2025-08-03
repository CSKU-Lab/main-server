package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SornchaiTheDev/cs-lab-backend/configs"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/services"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/middlewares"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/rest"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/sqlx"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/adapters/storage"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	// will be implemented in graceful shutdown
	ctx := context.TODO()

	minio := storage.NewMinIOStorage(ctx, config)

	userRepo := sqlx.NewSqlxUserRepository(db)
	userService := services.NewUserService(userRepo)

	userGroupRepo := sqlx.NewUserGroupRepository(db)
	userGroupService := services.NewUserGroupService(userGroupRepo)

	refreshTokenRepo := sqlx.NewSQLxRefreshTokenRepository(db)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	semesterRepo := sqlx.NewSqlxSemesterRepository(db)
	semesterService := services.NewSemesterService(semesterRepo)

	courseRepo := sqlx.NewSqlxCourseRepository(db)
	courseService := services.NewCourseService(courseRepo)

	sectionRepo := sqlx.NewSectionRepository(db)
	sectionService := services.NewSectionService(sectionRepo, courseRepo, minio)

	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
	})

	api := app.Group("/api/v1")

	rest.NewAuthRouter(api, config, userService, refreshTokenService)

	protectedApi := api.Group("/", middlewares.ProtectedRouteMiddleware(config.JWTSecret))

	rest.NewAdminRouter(&rest.AdminRouter{
		Router:           protectedApi,
		UserService:      userService,
		SemesterService:  semesterService,
		CourseService:    courseService,
		UserGroupService: userGroupService,
	})

	rest.NewCMSRouter(&rest.CMSRouter{
		Router:          protectedApi,
		SemesterService: semesterService,
		CourseService:   courseService,
		SectionService:  sectionService,
	})

	port := fmt.Sprintf(":%v", config.Port)

	err := app.Listen(port)
	if err != nil {
		log.Fatal("Error starting server on Port ", port)
	}

}
