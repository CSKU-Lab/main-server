package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type TestCase struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Input    string  `json:"input"`
	Output   string  `json:"output"`
	Message  string  `json:"message"`
	WallTime float32 `json:"wall_time"`
	Memory   int32   `json:"memory"`
}

type TestCaseGroup struct {
	ID        string     `json:"id"`
	Score     int        `json:"score,omitempty"`
	TestCases []TestCase `json:"test_cases"`
}

type TestCaseGroups []TestCaseGroup

func (t TestCaseGroups) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *TestCaseGroups) Scan(src any) error {
	var data []byte
	switch src := src.(type) {
	case []byte:
		data = src
	case string:
		data = []byte(src)
	default:
		return errors.New("unsupported data type for TestCaseGroups")
	}

	return json.Unmarshal(data, t)
}
