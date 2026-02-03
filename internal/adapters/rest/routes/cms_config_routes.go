package routes

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/gofiber/fiber/v3"
)

func NewCMSConfigRoutes(router fiber.Router, configGRPCClient configPB.ConfigServiceClient) {
	configRouter := router.Group("/configs")

	configRouter.Get("/runners", func(c fiber.Ctx) error {
		includeScriptQuery := c.Query("include_script", "false")

		runners, err := configGRPCClient.GetRunners(c.RequestCtx(), &configPB.GetRunnersRequest{
			IncludeName: true,
		})
		if err != nil {
			return err
		}

		if includeScriptQuery == "true" {
			var runnerConfigs []models.RunnerConfigDetail
			for _, runner := range runners.Runners {
				runnerConfigs = append(runnerConfigs, models.RunnerConfigDetail{
					RunnerConfig: models.RunnerConfig{
						ID:   runner.GetId(),
						Name: runner.GetName(),
					},
					BuildScript: runner.GetBuildScript(),
					RunScript:   runner.GetRunScript(),
				})
			}
			return c.JSON(runnerConfigs)
		}

		var runnerConfigs []models.RunnerConfig
		for _, runner := range runners.Runners {
			runnerConfigs = append(runnerConfigs, models.RunnerConfig{
				ID:   runner.GetId(),
				Name: runner.GetName(),
			})
		}

		return c.JSON(runnerConfigs)
	})

	configRouter.Get("/compare-scripts", func(c fiber.Ctx) error {
		compares, err := configGRPCClient.GetCompares(c.RequestCtx(), nil)
		if err != nil {
			return err
		}

		compareConfigs := make([]models.CompareConfig, 0, len(compares.Compares))
		for _, compare := range compares.Compares {
			compareConfigs = append(compareConfigs, models.CompareConfig{
				ID:   compare.GetId(),
				Name: compare.GetName(),
			})
		}

		return c.JSON(compareConfigs)
	})
}

// fiber:context-methods migrated
