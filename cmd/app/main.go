package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	"github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/adapters/storage/minio"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// temporaly way to just make it work for openhouse :D and need to clean this later
type RunnerConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RunnerConfigDetail struct {
	RunnerConfig
	BuildScript string `json:"build_script"`
	RunScript   string `json:"run_script"`
}

type RunExecutionRequest struct {
	Code     string `json:"code"`
	Input    string `json:"input"`
	RunnerID string `json:"runner_id"`
}

type CompareConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	config := configs.NewConfig()

	db := configs.NewDB(config)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	graderGRPCClient, closeConn, err := initGraderGRPCClient(config.GRADER_SERVER_URL)
	if err != nil {
		log.Fatal("Failed to connect to grader gRPC server: ", err)
	}
	defer closeConn()

	configGRPCClient, closeConn, err := initConfigGRPCClient(config.CONFIG_SERVER_URL)
	if err != nil {
		log.Fatal("Failed to connect to config gRPC server: ", err)
	}
	defer closeConn()

	taskGrpcClient, closeConn, err := initTaskGRPCClient(config.TASK_SERVER_URL)
	if err != nil {
		log.Fatal("Failed to connect to task gRPC server: ", err)
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

	courseRepo := sqlx.NewCourseRepository(db)
	courseCreatorRepo := sqlx.NewCourseCreatorRepository(db)
	courseService := services.NewCourseService(courseRepo, courseCreatorRepo, uowRepo)

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
	codeMaterial := registrables.NewCodeMaterial(codeMaterialRepo, taskGrpcClient, configGRPCClient, graderGRPCClient)
	materialRegistry.Register("code", codeMaterial)

	materialService := services.NewMaterialService(materialRepo, readMaterialTagRepo, uowRepo, userRepo, materialRegistry)

	affectedEntitiesFactory := registries.NewAffectedEntityFactory()
	deletedCourseAffected := registrables.NewDeletedCourseAffected(courseRepo, sectionRepo)
	deletedSemesterAffected := registrables.NewDeletedSemesterAffected(semesterRepo, sectionRepo, courseRepo)

	affectedEntitiesFactory.Register("course", deletedCourseAffected)
	affectedEntitiesFactory.Register("semester", deletedSemesterAffected)

	affectedEntitiesService := services.NewAffectedEntitiesService(affectedEntitiesFactory)

	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	api.Get("cms/config/runners", func(c *fiber.Ctx) error {
		includeScriptQuery := c.Query("include_script", "false")

		runners, err := configGRPCClient.GetRunners(c.Context(), &configPB.GetRunnersRequest{
			IncludeName: true,
		})
		if err != nil {
			return err
		}

		if includeScriptQuery == "true" {
			var runnerConfigs []RunnerConfigDetail
			for _, runner := range runners.Runners {
				runnerConfigs = append(runnerConfigs, RunnerConfigDetail{
					RunnerConfig: RunnerConfig{
						ID:   runner.GetId(),
						Name: runner.GetName(),
					},
					BuildScript: runner.GetBuildScript(),
					RunScript:   runner.GetRunScript(),
				})
			}
			return c.JSON(runnerConfigs)
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

	api.Get("/cms/config/compare-scripts", func(c *fiber.Ctx) error {
		compares, err := configGRPCClient.GetCompares(c.Context(), nil)
		if err != nil {
			return err
		}

		var compareConfigs []CompareConfig
		for _, compare := range compares.Compares {
			compareConfigs = append(compareConfigs, CompareConfig{
				ID:   compare.GetId(),
				Name: compare.GetName(),
			})
		}

		return c.JSON(compareConfigs)
	})

	api.Post("/playground/execute", func(c *fiber.Ctx) error {
		var req RunExecutionRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if req.RunnerID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "runner_id is required")
		}

		runReq := &graderPB.RunRequest{
			Input: req.Input,
			Files: []*graderPB.File{
				{
					Name:    "main.py",
					Content: req.Code,
				},
			},
			RunnerId: req.RunnerID,
		}

		stream, err := graderGRPCClient.Run(c.Context(), runReq)
		if err != nil {
			return err
		}

		c.Set(fiber.HeaderContentType, "text/event-stream")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		marshaler := protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseProtoNames:   true,
		}

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			writeEvent := func(event string, payload []byte) bool {
				if event != "" {
					if _, err := w.WriteString("event: " + event + "\n"); err != nil {
						return false
					}
				}
				if _, err := w.WriteString("data: "); err != nil {
					return false
				}
				if _, err := w.Write(payload); err != nil {
					return false
				}
				if _, err := w.WriteString("\n\n"); err != nil {
					return false
				}
				return w.Flush() == nil
			}

			for {
				result, err := stream.Recv()
				if err == io.EOF {
					writeEvent("done", []byte("{}"))
					return
				}
				if err != nil {
					log.Printf("grader stream error: %v", err)
					writeEvent("error", []byte(fmt.Sprintf("{\"error\":%q}", err.Error())))
					return
				}

				payload, err := marshaler.Marshal(result)
				if err != nil {
					log.Printf("marshal stream result error: %v", err)
					writeEvent("error", []byte(fmt.Sprintf("{\"error\":%q}", err.Error())))
					return
				}

				if ok := writeEvent("result", payload); !ok {
					return
				}
			}
		})

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
		Router:                  protectedApi,
		UserService:             userService,
		SemesterService:         semesterService,
		CourseService:           courseService,
		SectionService:          sectionService,
		MaterialService:         materialService,
		AffectedEntitiesService: affectedEntitiesService,
	})

	port := fmt.Sprintf(":%v", config.Port)

	go func() {
		<-ctx.Done()

		log.Println("Received shutdown signal, shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	err = app.Listen(port)
	if err != nil {
		log.Fatal("Error starting server on Port ", port, ": ", err)
	}

	log.Println("Server stopped")
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

func initTaskGRPCClient(clientAddr string) (taskPB.TaskServiceClient, func(), error) {
	conn, err := grpc.NewClient(clientAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}

	client := taskPB.NewTaskServiceClient(conn)

	return client, func() {
		conn.Close()
	}, nil
}
