package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/gofiber/fiber/v3"
)

func NewCMSConfigRoutes(router fiber.Router, configGRPCClient configPB.ConfigServiceClient) {
	configRouter := router.Group("/configs", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	configRouter.Post("/runners", middlewares.ValidateMiddleware[requests.CreateRunnerRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateRunnerRequest)
		runner, err := configGRPCClient.CreateRunner(c.RequestCtx(), &configPB.CreateRunnerRequest{
			Name:        req.Name,
			BuildScript: req.BuildScript,
			RunScript:   req.RunScript,
		})
		if err != nil {
			return err
		}
		return c.JSON(runner)
	})

	configRouter.Patch("/runners/:id", middlewares.ValidateMiddleware[requests.UpdateRunnerRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.UpdateRunnerRequest)
		id := c.Params("id")
		runner, err := configGRPCClient.UpdateRunner(c.RequestCtx(), &configPB.UpdateRunnerRequest{
			Id:          id,
			Name:        req.Name,
			BuildScript: req.BuildScript,
			RunScript:   req.RunScript,
		})
		if err != nil {
			return err
		}
		return c.JSON(runner)
	})

	configRouter.Delete("/runners/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		_, err := configGRPCClient.DeleteRunner(c.RequestCtx(), &configPB.DeleteRunnerRequest{
			Id: id,
		})
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"message": "Runner deleted successfully",
		})
	})

	configRouter.Get("/runners", func(c fiber.Ctx) error {
		includeScriptQuery := c.Query("include_script", "false")
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "20")
		sortOrder := c.Query("sort_order", "desc")
		search := c.Query("search", "")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}

		paginationRes, err := configGRPCClient.GetRunnersPagination(c.RequestCtx(), &configPB.GetRunnersPaginationRequest{
			IncludeName: true,
			Pagination: &configPB.PaginationRequest{
				PageSize:  int32(pageSize),
				Page:      int32(page),
				SortOrder: sortOrder,
				Search:    search,
			},
		})
		if err != nil {
			return err
		}

		if includeScriptQuery == "true" {
			var runnerConfigs []models.RunnerConfigDetail
			for _, runner := range paginationRes.Runners {
				runnerConfigs = append(runnerConfigs, models.RunnerConfigDetail{
					RunnerConfig: models.RunnerConfig{
						ID:   runner.GetId(),
						Name: runner.GetName(),
					},
					BuildScript: runner.GetBuildScript(),
					RunScript:   runner.GetRunScript(),
				})
			}
			return c.JSON(fiber.Map{
				"pagination": fiber.Map{
					"page":       page,
					"total_page": int(math.Ceil(float64(paginationRes.Count) / float64(pageSize))),
					"total_rows": paginationRes.Count,
				},
				"data": runnerConfigs,
			})
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": int(math.Ceil(float64(paginationRes.Count) / float64(pageSize))),
				"total_rows": paginationRes.Count,
			},
			"data": paginationRes.Runners,
		})
	})

	configRouter.Get("/compare-scripts", func(c fiber.Ctx) error {
		pageQuery := c.Query("page", "1")
		pageSizeQuery := c.Query("page_size", "20")
		sortOrder := c.Query("sort_order", "desc")
		search := c.Query("search", "")

		page, err := strconv.Atoi(pageQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page"})
		}

		pageSize, err := strconv.Atoi(pageSizeQuery)
		if err != nil {
			return cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid page size"})
		}

		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}

		paginationRes, err := configGRPCClient.GetComparesPagination(c.RequestCtx(), &configPB.GetComparesPaginationRequest{
			Pagination: &configPB.PaginationRequest{
				PageSize:  int32(pageSize),
				Page:      int32(page),
				SortOrder: sortOrder,
				Search:    search,
			},
		})
		if err != nil {
			return err
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": int(math.Ceil(float64(paginationRes.Count) / float64(pageSize))),
				"total_rows": paginationRes.Count,
			},
			"data": paginationRes.Compares,
		})
	})

	configRouter.Post("/compare-scripts", middlewares.ValidateMiddleware[requests.CreateCompareRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateCompareRequest)
		compare, err := configGRPCClient.CreateCompare(c.RequestCtx(), &configPB.CreateCompareRequest{
			Name:        req.Name,
			RunScript:   req.RunScript,
			RunName:     req.RunName,
			Description: req.Description,
			BuildScript: req.BuildScript,
			Files:       requests.MapCompareFilesToPB(req.Files),
		})
		if err != nil {
			return err
		}
		return c.JSON(compare)
	})

	configRouter.Patch("/compare-scripts/:id", middlewares.ValidateMiddleware[requests.UpdateCompareRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.UpdateCompareRequest)
		id := c.Params("id")
		compare, err := configGRPCClient.UpdateCompare(c.RequestCtx(), &configPB.UpdateCompareRequest{
			Id:          id,
			Name:        req.Name,
			RunScript:   req.RunScript,
			RunName:     req.RunName,
			Description: req.Description,
			BuildScript: req.BuildScript,
			Files:       requests.MapCompareFilesToPB(req.Files),
		})
		if err != nil {
			return err
		}
		return c.JSON(compare)
	})

	configRouter.Delete("/compare-scripts/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		_, err := configGRPCClient.DeleteCompare(c.RequestCtx(), &configPB.DeleteCompareRequest{
			Id: id,
		})
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"message": "Compare deleted successfully",
		})
	})
}

// fiber:context-methods migrated
