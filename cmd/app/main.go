package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/adapters/storage/minio"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// temporaly way to just make it work for openhouse :D and need to clean this later
type RunnerConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RunExecutionRequest struct {
	Code     string `json:"code"`
	Input    string `json:"input"`
	RunnerID string `json:"runner_id"`
}

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	// will be implemented in graceful shutdown
	ctx := context.TODO()

	graderClient, closeConn, err := initGraderGRPCClient(config.GRADER_SERVER_URL)
	if err != nil {
		log.Fatal("Failed to connect to grader gRPC server: ", err)
	}
	defer closeConn()

	configClient, closeConn, err := initConfigGRPCClient(config.CONFIG_SERVER_URL)
	if err != nil {
		log.Fatal("Failed to connect to config gRPC server: ", err)
	}
	defer closeConn()

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

	materialRepo := sqlx.NewMaterialRepository(db)
	readMaterialTagRepo := sqlx.NewReadMaterialTagRepository(db)
	codeMaterialRepo := sqlx.NewCodeMaterialRepository(db)

	materialRegistry := registries.NewMaterialRegistry()
	codeMaterial := registrables.NewCodeMaterial(codeMaterialRepo)
	materialRegistry.Register("code", codeMaterial)

	materialService := services.NewMaterialService(materialRepo, readMaterialTagRepo, uowRepo, userRepo, materialRegistry)

	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	api.Get("/config/runners", func(c *fiber.Ctx) error {
		runners, err := configClient.GetRunners(c.Context(), &configPB.GetRunnersRequest{
			IncludeName: true,
		})
		if err != nil {
			return err
		}

		var runnerConfigs []RunnerConfig
		for _, runner := range runners.Runners {
			runnerConfigs = append(runnerConfigs, RunnerConfig{
				ID:   runner.GetId(),
				Name: runner.GetName(),
			})
		}

		return c.JSON(runnerConfigs)

	})

	api.Post("/execute", func(c *fiber.Ctx) error {
		var req RunExecutionRequest
		err := c.BodyParser(&req)
		if err != nil {
			return err
		}

		stream, err := graderClient.Run(c.Context(), &graderPB.RunRequest{
			Input: req.Input,
			Files: []*graderPB.File{
				{
					Name:    "main.py",
					Content: req.Code,
				},
			},
			RunnerId: req.RunnerID,
		})
		if err != nil {
			return err
		}

		for {
			result, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			return c.JSON(fiber.Map{
				"executionID": result.ExecutionId,
				"status":      result.Status.String(),
				"output":      result.Output,
				"wall_time":   result.WallTime,
				"memory":      result.Memory,
			})
		}
		return nil
	})

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
		MaterialService: materialService,
	})

	port := fmt.Sprintf(":%v", config.Port)

	err = app.Listen(port)
	if err != nil {
		log.Fatal("Error starting server on Port ", port)
	}

}

func initConfigGRPCClient(clientAddr string) (configPB.ConfigServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := configPB.NewConfigServiceClient(conn)

	return client, func() {
		conn.Close()
	}, nil
}

func initGraderGRPCClient(clientAddr string) (graderPB.GraderServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := graderPB.NewGraderServiceClient(conn)

	return client, func() {
		conn.Close()
	}, nil
}
