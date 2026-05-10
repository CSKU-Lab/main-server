package rest

import (
	"bufio"
	"fmt"
	"io"
	"log"

	graderPB "github.com/CSKU-Lab/main-server/genproto/grader/v1"
	"github.com/gofiber/fiber/v3"
	"google.golang.org/protobuf/encoding/protojson"
)

type TmpFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type RunExecutionRequest struct {
	Files    []TmpFile `json:"files"`
	Input    string    `json:"input"`
	RunnerID string    `json:"runner_id"`
}

type PlaygroundHandler struct {
	graderClient graderPB.GraderServiceClient
}

func NewPlaygroundHandler(graderClient graderPB.GraderServiceClient) *PlaygroundHandler {
	return &PlaygroundHandler{graderClient: graderClient}
}

func (h *PlaygroundHandler) Execute(c fiber.Ctx) error {
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

	stream, err := h.graderClient.Run(c.RequestCtx(), runReq)
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
}
