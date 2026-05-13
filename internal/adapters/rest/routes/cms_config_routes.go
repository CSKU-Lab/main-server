package routes

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	"github.com/CSKU-Lab/main-server/internal/adapters/middlewares"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/queue"
	"github.com/gofiber/fiber/v3"
)

func NewCMSConfigRoutes(router fiber.Router, configGRPCClient configPB.ConfigServiceClient, q queue.Queue) {
	// Base config router - requires authentication but no specific role
	// Individual routes will have their own permission checks
	configRouter := router.Group("/configs", middlewares.RBACMiddleware([]models.Role{
		models.ADMIN,
		models.INSTRUCTOR,
	}))

	// Runner routes - viewable by admin and instructor, modifiable by admin only
	runnerRouter := configRouter.Group("/runners")

	// GET /configs/runners/:id - View (Admin or Instructor)
	runnerRouter.Get("/:id", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		id := c.Params("id")
		runner, err := configGRPCClient.GetRunner(c.RequestCtx(), &configPB.GetRunnerRequest{
			Id: id,
		})
		if err != nil {
			return err
		}
		return c.JSON(&models.RunnerConfigDetail{
			RunnerConfig: &models.RunnerConfig{
				ID:          runner.GetId(),
				Name:        runner.GetName(),
				Description: runner.GetDescription(),
			},
			BuildScript:  runner.GetBuildScript(),
			RunScript:    runner.GetRunScript(),
			InitialFiles: pbFilesToModelFiles(runner.GetInitialFiles()),
		})
	})

	runnerRouter.Post("/", middlewares.RequireAdmin(), middlewares.ValidateMiddleware[requests.CreateRunnerRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateRunnerRequest)
		runner, err := configGRPCClient.CreateRunner(c.RequestCtx(), &configPB.CreateRunnerRequest{
			Name:        req.Name,
			Description: req.Description,
		})
		if err != nil {
			return err
		}
		return c.JSON(runner)
	})

	runnerRouter.Patch("/:id", middlewares.RequireAdmin(), middlewares.ValidateMiddleware[requests.UpdateRunnerRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.UpdateRunnerRequest)
		id := c.Params("id")
		payload := &configPB.UpdateRunnerRequest{
			Id:          id,
			Name:        req.Name,
			Description: req.Description,
			BuildScript: req.BuildScript,
			RunScript:   req.RunScript,
		}

		if req.InitialFiles != nil {
			payload.InitialFiles = requests.MapConfigFilesToPB(*req.InitialFiles)
		}

		runner, err := configGRPCClient.UpdateRunner(c.RequestCtx(), payload)
		if err != nil {
			return err
		}

		return c.JSON(runner)
	})

	runnerRouter.Delete("/:id", middlewares.RequireAdmin(), func(c fiber.Ctx) error {
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

	type runnerTestRequest struct {
		InitialFiles []models.ConfigFile `json:"initial_files"`
		Input        string              `json:"input"`
		RunScript    string              `json:"run_script"`
		BuildScript  string              `json:"build_script"`
	}

	runnerRouter.Post("/:id/test", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		id := c.Params("id")
		body := requests.TestRunnerRequest{}

		err := c.Bind().JSON(&body)
		if err != nil {
			return err
		}

		qName, err := q.CreateQueue(c.RequestCtx(), "runner_test:"+id, &queue.QueueOptions{
			AutoDelete: true,
			Exclusive:  true,
		})

		payloadBytes, err := json.Marshal(&runnerTestRequest{
			InitialFiles: reqInitialFilesToModel(body.InitialFiles),
			Input:        body.Input,
			RunScript:    body.RunScript,
			BuildScript:  body.BuildScript,
		})
		if err != nil {
			return err
		}

		err = q.Publish(c.RequestCtx(), "", "runner_test", &queue.Derivery{
			CorrelationID: id,
			ReplyTo:       qName,
			Body:          payloadBytes,
		})
		if err != nil {
			return err
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		ctx, cancel := context.WithCancel(context.Background())

		c.Status(fiber.StatusOK).RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
			fmt.Fprint(w, "event: connected\n\n")
			err := w.Flush()
			if err != nil {
				cancel()
				log.Println(err)
				return
			}

			err = q.Consume(ctx, qName, 1, true, func(derivery *queue.Derivery, exit chan struct{}) error {
				runResult := &models.TestRunnerResult{}
				err := json.Unmarshal(derivery.Body, &runResult)
				if err != nil {
					return err
				}

				runResultBytes, err := json.Marshal(runResult)
				if err != nil {
					return err
				}

				fmt.Fprintf(w, "data: %s\n\n", runResultBytes)
				err = w.Flush()
				if err != nil {
					cancel()
					log.Println(err)
					return err

				}

				if runResult.Status != models.CODE_EXECUTION_QUEUED && runResult.Status != models.CODE_EXECUTION_RUNNING {
					exit <- struct{}{}
				}

				return nil
			})
			if err != nil {
				log.Println("Error consuming runner test result:", err)
				return
			}

			fmt.Fprint(w, "event: done\n\n")
			err = w.Flush()
			if err != nil {
				cancel()
				log.Println(err)
				return
			}
		})
		return nil
	})

	runnerRouter.Get("/", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		includeScriptsQuery := c.Query("include_scripts", "false")
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
			Pagination: &configPB.PaginationRequest{
				PageSize:  int32(pageSize),
				Page:      int32(page),
				SortOrder: sortOrder,
				Search:    search,
			},
			IncludeScripts: includeScriptsQuery == "true",
		})
		if err != nil {
			return err
		}

		var data any
		if includeScriptsQuery == "true" {
			runnerConfigs := make([]models.RunnerConfigDetail, len(paginationRes.Runners))
			for i, runner := range paginationRes.Runners {
				runnerConfigs[i] = models.RunnerConfigDetail{
					RunnerConfig: &models.RunnerConfig{
						ID:          runner.GetId(),
						Name:        runner.GetName(),
						Description: runner.GetDescription(),
					},
					BuildScript:  runner.GetBuildScript(),
					RunScript:    runner.GetRunScript(),
					InitialFiles: pbFilesToModelFiles(runner.GetInitialFiles()),
				}
			}
			data = runnerConfigs
		} else {
			runnerConfigs := make([]models.RunnerConfig, len(paginationRes.Runners))
			for i, runner := range paginationRes.Runners {
				runnerConfigs[i] = models.RunnerConfig{
					ID:          runner.GetId(),
					Name:        runner.GetName(),
					Description: runner.GetDescription(),
				}
			}
			data = runnerConfigs
		}

		return c.JSON(fiber.Map{
			"pagination": fiber.Map{
				"page":       page,
				"total_page": int(math.Ceil(float64(paginationRes.Count) / float64(pageSize))),
				"total_rows": paginationRes.Count,
			},
			"data": data,
		})
	})

	type compareTestRequest struct {
		Files       []models.ConfigFile `json:"files"`
		BuildScript string              `json:"build_script"`
		RunScript   string              `json:"run_script"`
		SolOutput   string              `json:"sol_output"`
		Output      string              `json:"output"`
	}

	type compareTestPayload struct {
		Files       []models.ConfigFile `json:"files"`
		BuildScript string              `json:"build_script"`
		RunScript   string              `json:"run_script"`
		RunName     string              `json:"run_name"`
		SolOutput   string              `json:"sol_output"`
		Output      string              `json:"output"`
	}

	configRouter.Post("/compare-scripts/:id/test", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		id := c.Params("id")
		body := compareTestRequest{}
		if err := c.Bind().JSON(&body); err != nil {
			return err
		}

		compare, err := configGRPCClient.GetCompare(c.RequestCtx(), &configPB.GetCompareRequest{
			Id: id,
		})
		if err != nil {
			return err
		}

		qName, err := q.CreateQueue(c.RequestCtx(), "compare_test:"+id, &queue.QueueOptions{
			AutoDelete: true,
			Exclusive:  true,
		})
		if err != nil {
			return err
		}

		payloadBytes, err := json.Marshal(&compareTestPayload{
			Files:       body.Files,
			BuildScript: body.BuildScript,
			RunScript:   body.RunScript,
			RunName:     compare.GetRunName(),
			SolOutput:   body.SolOutput,
			Output:      body.Output,
		})
		if err != nil {
			return err
		}

		err = q.Publish(c.RequestCtx(), "", "compare_test", &queue.Derivery{
			CorrelationID: id,
			ReplyTo:       qName,
			Body:          payloadBytes,
		})
		if err != nil {
			return err
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		ctx, cancel := context.WithCancel(context.Background())

		c.Status(fiber.StatusOK).RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
			fmt.Fprint(w, "event: connected\n\n")
			err := w.Flush()
			if err != nil {
				cancel()
				log.Println(err)
				return
			}

			err = q.Consume(ctx, qName, 1, true, func(derivery *queue.Derivery, exit chan struct{}) error {
				runResult := &models.TestRunnerResult{}
				err := json.Unmarshal(derivery.Body, &runResult)
				if err != nil {
					return err
				}

				runResultBytes, err := json.Marshal(runResult)
				if err != nil {
					return err
				}

				fmt.Fprintf(w, "data: %s\n\n", runResultBytes)
				err = w.Flush()
				if err != nil {
					cancel()
					log.Println(err)
					return err
				}

				if runResult.Status != models.CODE_EXECUTION_QUEUED && runResult.Status != models.CODE_EXECUTION_RUNNING {
					exit <- struct{}{}
				}

				return nil
			})
			if err != nil {
				log.Println("Error consuming compare test result:", err)
				return
			}

			fmt.Fprint(w, "event: done\n\n")
			err = w.Flush()
			if err != nil {
				cancel()
				log.Println(err)
				return
			}
		})
		return nil
	})

	// GET /configs/compare-scripts/:id - View (Admin or Instructor)
	configRouter.Get("/compare-scripts/:id", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
		id := c.Params("id")
		compare, err := configGRPCClient.GetCompare(c.RequestCtx(), &configPB.GetCompareRequest{
			Id: id,
		})
		if err != nil {
			return err
		}
		return c.JSON(&models.CompareConfigDetail{
			CompareConfig: &models.CompareConfig{
				ID:          compare.GetId(),
				Name:        compare.GetName(),
				Description: compare.GetDescription(),
			},
			BuildScript: compare.GetBuildScript(),
			RunScript:   compare.GetRunScript(),
			RunName:     compare.GetRunName(),
			Files:       pbFilesToModelFiles(compare.GetFiles()),
		})
	})

	configRouter.Get("/compare-scripts", middlewares.RequireAdminOrInstructor(), func(c fiber.Ctx) error {
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

	compares := make([]*configPB.CompareResponse, len(paginationRes.Compares))
	copy(compares, paginationRes.Compares)

	return c.JSON(fiber.Map{
		"pagination": fiber.Map{
			"page":       page,
			"total_page": int(math.Ceil(float64(paginationRes.Count) / float64(pageSize))),
			"total_rows": paginationRes.Count,
		},
		"data": compares,
	})
	})

	configRouter.Post("/compare-scripts", middlewares.RequireAdmin(), middlewares.ValidateMiddleware[requests.CreateCompareRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.CreateCompareRequest)
		compare, err := configGRPCClient.CreateCompare(c.RequestCtx(), &configPB.CreateCompareRequest{
			Name:        req.Name,
			Description: req.Description,
			BuildScript: "#!/bin/bash\n",
			RunScript:   "#!/bin/bash\n",
			RunName:     "main",
			Files: []*configPB.File{
				{
					Name: "GUIDE.md",
					Content: "# Compare Script Guide\n\n" +
						"## What is a Compare Script?\n\n" +
						"A compare script is a program that receives two text outputs and decides\n" +
						"whether they match — using any logic you want (exact match, prefix match,\n" +
						"numeric tolerance, etc.).\n\n" +
						"## Result convention\n\n" +
						"| `/box/compare_result.txt` content | Exit code | Verdict |\n" +
						"|-----------------------------------|-----------|----------|\n" +
						"| empty (or file absent) | 0 | **Accepted** |\n" +
						"| non-empty message | 1 | **Wrong Answer** — message shown to user |\n\n" +
						"Write your verdict message to `/box/compare_result.txt` and exit with the correct code.\n" +
						"Anything printed to stdout is shown in the **Test** panel (useful for debugging) but ignored during grading.\n\n" +
						"## Environment variables\n\n" +
						"| Variable | Description |\n" +
						"|----------|-------------|\n" +
						"| `$RUN_NAME` | Absolute path to your compare script inside the sandbox |\n" +
						"| `$SOL_OUTPUT` | Path to file containing the expected (solution) output |\n" +
						"| `$OUTPUT` | Path to file containing the student's output |\n\n" +
						"## Example run_script.sh\n\n" +
						"```bash\n" +
						"#!/bin/bash\n" +
						"python3 \"$RUN_NAME\"\n" +
						"```\n\n" +
						"## Example compare script (exact_match.py)\n\n" +
						"```python\n" +
						"import os, sys\n\n" +
						"sol = open(os.environ[\"SOL_OUTPUT\"]).read().strip()\n" +
						"out = open(os.environ[\"OUTPUT\"]).read().strip()\n\n" +
						"if sol != out:\n" +
						"    with open(\"/box/compare_result.txt\", \"w\") as f:\n" +
						"        f.write(f\"Expected: {sol}\\nGot:      {out}\\n\")\n" +
						"    sys.exit(1)\n" +
						"# exit 0, no file written = Accepted\n" +
						"```\n\n" +
						"## Typical workflow\n\n" +
						"1. Add your compare script (e.g. `exact_match.py`) to this folder.\n" +
						"2. Set **Run Name** to that filename (`exact_match.py`).\n" +
						"3. Write `build_script.sh` to compile anything if needed.\n" +
						"4. Write `run_script.sh` to invoke `$RUN_NAME`.\n" +
						"5. Use the **Test** tab to verify behaviour before assigning to a task.\n",
				},
			},
		})
		if err != nil {
			return err
		}
		return c.JSON(compare)
	})

	configRouter.Patch("/compare-scripts/:id", middlewares.RequireAdmin(), middlewares.ValidateMiddleware[requests.UpdateCompareRequest](), func(c fiber.Ctx) error {
		req := c.Locals("body").(*requests.UpdateCompareRequest)
		id := c.Params("id")
		compare, err := configGRPCClient.UpdateCompare(c.RequestCtx(), &configPB.UpdateCompareRequest{
			Id:          id,
			Name:        req.Name,
			RunScript:   req.RunScript,
			RunName:     req.RunName,
			Description: req.Description,
			BuildScript: req.BuildScript,
			Files:       requests.MapConfigFilesToPB(req.Files),
		})
		if err != nil {
			return err
		}
		return c.JSON(compare)
	})

	configRouter.Delete("/compare-scripts/:id", middlewares.RequireAdmin(), func(c fiber.Ctx) error {
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

func pbFilesToModelFiles(pbFile []*configPB.File) []models.ConfigFile {
	files := make([]models.ConfigFile, len(pbFile))
	for i, file := range pbFile {
		files[i] = models.ConfigFile{
			Name:    file.GetName(),
			Content: file.GetContent(),
		}
	}
	return files
}

func modelFilesToPBFiles(modelFiles []models.ConfigFile) []*configPB.File {
	files := make([]*configPB.File, len(modelFiles))
	for i, file := range modelFiles {
		files[i] = &configPB.File{
			Name:    file.Name,
			Content: file.Content,
		}
	}
	return files
}

func reqInitialFilesToModel(files []requests.ConfigFile) []models.ConfigFile {
	modelFiles := make([]models.ConfigFile, len(files))
	for i, file := range files {
		modelFiles[i] = models.ConfigFile{
			Name:    file.Name,
			Content: file.Content,
		}
	}
	return modelFiles
}
