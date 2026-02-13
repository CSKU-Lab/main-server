package routes

import (
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/gofiber/fiber/v3"
)

func NewCMSConfigRoutes(router fiber.Router, configGRPCClient configPB.ConfigServiceClient) {
	configRouter := router.Group("/configs")

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
}

// fiber:context-methods migrated
