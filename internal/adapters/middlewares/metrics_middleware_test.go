package middlewares

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
)

// observedCount returns how many requests were recorded for the given labels
// without creating the series as a side effect.
func observedCount(t *testing.T, method, route, status string) uint64 {
	t.Helper()

	ch := make(chan prometheus.Metric, 64)
	go func() {
		httpRequestDuration.Collect(ch)
		close(ch)
	}()

	want := map[string]string{"method": method, "route": route, "status": status}
	var count uint64
	for m := range ch {
		var pb dto.Metric
		if err := m.Write(&pb); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		matched := 0
		for _, lp := range pb.GetLabel() {
			if want[lp.GetName()] == lp.GetValue() {
				matched++
			}
		}
		if matched == len(want) {
			count += pb.GetHistogram().GetSampleCount()
		}
	}
	return count
}

func TestMetricsMiddleware(t *testing.T) {
	errTeapot := errors.New("teapot")

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if errors.Is(err, errTeapot) {
				return c.SendStatus(fiber.StatusTeapot)
			}
			return fiber.DefaultErrorHandler(c, err)
		},
	})
	app.Use(MetricsMiddleware())
	app.Get("/labs/:id", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	app.Get("/fail", func(c fiber.Ctx) error {
		return errTeapot
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantRoute  string
		wantClass  string
	}{
		{name: "uses route template, not raw path", path: "/labs/42", wantStatus: fiber.StatusOK, wantRoute: "/labs/:id", wantClass: "2xx"},
		{name: "records status set by app error handler", path: "/fail", wantStatus: fiber.StatusTeapot, wantRoute: "/fail", wantClass: "4xx"},
		{name: "unmatched path does not create a per-path series", path: "/nope/123", wantStatus: fiber.StatusNotFound, wantRoute: "/", wantClass: "4xx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("response status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if got := observedCount(t, fiber.MethodGet, tt.wantRoute, tt.wantClass); got != 1 {
				t.Fatalf("observations for {route=%q, status=%q} = %d, want 1", tt.wantRoute, tt.wantClass, got)
			}
		})
	}

	if got := testutil.ToFloat64(httpRequestsInFlight); got != 0 {
		t.Fatalf("in-flight gauge = %v after requests finished, want 0", got)
	}
}
