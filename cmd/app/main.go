package main

import (
	"context"
	"fmt"
	"log"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/services"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/adapters/storage/minio"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	// will be implemented in graceful shutdown
	ctx := context.TODO()

	minio := minio.New(ctx, config)

	uowRepo := sqlx.NewUoWRepository(ctx, db)

	userRepo := sqlx.NewUserRepository(db)
	userPasswordRepo := sqlx.NewUserPasswordRepository(db)
	userGroupRepo := sqlx.NewUserGroupRepository(db)

	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)

	userGroupService := services.NewUserGroupService(userGroupRepo)

	refreshTokenRepo := sqlx.NewSQLxRefreshTokenRepository(db)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	courseRepo := sqlx.NewSqlxCourseRepository(db)
	courseCreatorRepo := sqlx.NewCourseCreatorRepository(db)
	courseService := services.NewCourseService(courseRepo, courseCreatorRepo)

	sectionRepo := sqlx.NewSectionRepository(db)
	sectionInstructorRepo := sqlx.NewSectionInstructorRepository(db)
	sectionStudentRepo := sqlx.NewSectionStudentRepository(db)
	sectionService := services.NewSectionService(config, sectionRepo, uowRepo, courseRepo, sectionInstructorRepo, sectionStudentRepo, minio, userRepo)

	semesterRepo := sqlx.NewSqlxSemesterRepository(db)
	semesterService := services.NewSemesterService(semesterRepo, sectionRepo, courseRepo)

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
