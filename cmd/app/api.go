package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/services"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/adapters/rest"
	sqlxAdapter "github.com/CSKU-Lab/main-server/internal/adapters/sqlx"
	"github.com/CSKU-Lab/main-server/internal/adapters/storage/minio"
)

// temporaly files
type TmpFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type RunExecutionRequest struct {
	Files    []TmpFile `json:"files"`
	Input    string    `json:"input"`
	RunnerID string    `json:"runner_id"`
}

func startApiServer(ctx context.Context, logger *zap.SugaredLogger, db *sqlx.DB, config *configs.Config) {
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

	rClient, err := pubsub.NewRedis(config.REDIS_SERVER_URL)
	if err != nil {
		logger.Fatalln(err)
	}

	minio := minio.New(ctx, config)

	uowRepo := sqlxAdapter.NewUoWRepository(ctx, db)

	userRepo := sqlxAdapter.NewUserRepository(db)
	userPasswordRepo := sqlxAdapter.NewUserPasswordRepository(db)
	userGroupRepo := sqlxAdapter.NewUserGroupRepository(db)

	userService := services.NewUserService(userRepo, userPasswordRepo, userGroupRepo, uowRepo)

	userGroupService := services.NewUserGroupService(userGroupRepo)

	refreshTokenRepo := sqlxAdapter.NewSQLxRefreshTokenRepository(db)
	refreshTokenService := services.NewRefreshTokenService(refreshTokenRepo)

	courseRepo := sqlxAdapter.NewCourseRepository(db)
	courseCreatorRepo := sqlxAdapter.NewCourseCreatorRepository(db)
	courseService := services.NewCourseService(courseRepo, courseCreatorRepo, uowRepo)

	sectionRepo := sqlxAdapter.NewSectionRepository(db)
	sectionInstructorRepo := sqlxAdapter.NewSectionInstructorRepository(db)
	sectionStudentRepo := sqlxAdapter.NewSectionStudentRepository(db)
	sectionLogRepo := sqlxAdapter.NewSectionLogRepository(db)

	semesterRepo := sqlxAdapter.NewSqlxSemesterRepository(db)
	semesterService := services.NewSemesterService(semesterRepo, sectionRepo, courseRepo)

	sectionLogService := services.NewSectionLogService(sectionLogRepo)
	sectionService := services.NewSectionService(config, sectionRepo, uowRepo, courseRepo, sectionInstructorRepo, sectionStudentRepo, minio, userRepo, semesterRepo, sectionLogService)

	materialRepo := sqlxAdapter.NewMaterialRepository(db)
	readMaterialTagRepo := sqlxAdapter.NewReadMaterialTagRepository(db)
	codeMaterialRepo := sqlxAdapter.NewCodeMaterialRepository(db)

	materialRegistry := registries.NewMaterialRegistry()
	codeMaterial := registrables.NewCodeMaterial(codeMaterialRepo, taskGrpcClient, configGRPCClient, graderGRPCClient)
	materialRegistry.Register("code", codeMaterial)

	materialService := services.NewMaterialService(materialRepo, readMaterialTagRepo, uowRepo, userRepo, materialRegistry)
	materialAssetService := services.NewMaterialAssetService(config, minio)

	labRepo := sqlxAdapter.NewSqlxLabRepository(db)
	labService := services.NewLabService(labRepo, courseRepo, uowRepo)

	labSectionRepo := sqlxAdapter.NewSqlxLabSectionRepository(db)
	labSectionService := services.NewLabSectionService(labSectionRepo, uowRepo, labRepo, sectionRepo, sectionStudentRepo)

	labMaterialRepo := sqlxAdapter.NewSqlxLabMaterialRepository(db)
	labMaterialService := services.NewLabMaterialService(labMaterialRepo, uowRepo, labRepo, materialRepo, readMaterialTagRepo)

	defaultLabRepo := sqlxAdapter.NewSqlxDefaultLabRepository(db)
	defaultLabService := services.NewDefaultLabService(defaultLabRepo, uowRepo, courseRepo, labRepo)

	affectedEntitiesFactory := registries.NewAffectedEntityFactory()
	deletedCourseAffected := registrables.NewDeletedCourseAffected(courseRepo, sectionRepo)
	deletedSemesterAffected := registrables.NewDeletedSemesterAffected(semesterRepo, sectionRepo, courseRepo)
	deletedSectionAffected := registrables.NewDeletedSectionAffected(sectionStudentRepo)
	deletedLabAffected := registrables.NewDeletedLabAffected(labRepo, labSectionRepo, labMaterialRepo, defaultLabRepo, uowRepo)
	deletedLabSectionAffected := registrables.NewDeletedLabSectionAffected()

	affectedEntitiesFactory.Register("course", deletedCourseAffected)
	affectedEntitiesFactory.Register("semester", deletedSemesterAffected)
	affectedEntitiesFactory.Register("section", deletedSectionAffected)
	affectedEntitiesFactory.Register("lab", deletedLabAffected)
	affectedEntitiesFactory.Register("lab_section", deletedLabSectionAffected)

	affectedEntitiesService := services.NewAffectedEntitiesService(affectedEntitiesFactory)

	errHandlerMiddleware := middlewares.NewErrorHandlerMiddleware(config)

	sidebarService := services.NewSidebarService(courseRepo, sectionStudentRepo, labSectionRepo, labMaterialRepo)

	submissionRepo := sqlxAdapter.NewSubmissionRepository(db)
	codeSubmissionRepo := sqlxAdapter.NewCodeSubmission(db)

	codeSubmissionRegistrable := registrables.NewCodeSubmission(codeSubmissionRepo, codeMaterialRepo, submissionRepo)

	submissionRegistry := registries.NewSubmission()
	submissionRegistry.Register("code", codeSubmissionRegistrable)
	submissionService := services.NewSubmissionService(&services.SubmissionServiceArgs{
		SubmissionRepository:     submissionRepo,
		MaterialRepository:       materialRepo,
		UowRepository:            uowRepo,
		SubmissionRegistry:       submissionRegistry,
		SectionStudentRepository: sectionStudentRepo,
		PubSub:                   rClient,
	})

	app := fiber.New(fiber.Config{
		ErrorHandler: errHandlerMiddleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10 MB,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	api := app.Group("/api/v1")

	api.Post("/playground/execute", func(c fiber.Ctx) error {
		var req RunExecutionRequest
		if err := c.Bind().Body(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if req.RunnerID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "runner_id is required")
		}

		graderFiles := make([]*graderPB.File, 0, len(req.Files))
		for _, f := range req.Files {
			graderFiles = append(graderFiles, &graderPB.File{
				Name:    f.Name,
				Content: f.Content,
			})
		}

		runReq := &graderPB.RunRequest{
			Input:    req.Input,
			Files:    graderFiles,
			RunnerId: req.RunnerID,
		}

		stream, err := graderGRPCClient.Run(c.RequestCtx(), runReq)
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

		c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
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
		LabService:              labService,
		LabSectionService:       labSectionService,
		LabMaterialService:      labMaterialService,
		DefaultLabService:       defaultLabService,
		SectionLogService:       sectionLogService,
		MaterialAssetService:    materialAssetService,
		ConfigGRPCClient:        configGRPCClient,
	})

	rest.NewCoreRouter(&rest.CoreRouter{
		Router:                protectedApi,
		SectionService:        sectionService,
		LabSectionService:     labSectionService,
		LabService:            labService,
		SectionStudentService: sectionStudentRepo,
		LabMaterialService:    labMaterialService,
		CourseService:         courseService,
		SidebarService:        sidebarService,
		SubmissionService:     submissionService,
		PubSub:                rClient,
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

// fiber:context-methods migrated
