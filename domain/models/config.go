package models

type RunnerConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ConfigFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type RunnerConfigDetail struct {
	*RunnerConfig
	BuildScript  string       `json:"build_script"`
	RunScript    string       `json:"run_script"`
	InitialFiles []ConfigFile `json:"initial_files"`
}

type CompareConfig struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CompareConfigDetail struct {
	*CompareConfig
	BuildScript string       `json:"build_script"`
	RunScript   string       `json:"run_script"`
	RunName     string       `json:"run_name"`
	Files       []ConfigFile `json:"files"`
}

type TestRunnerResult struct {
	ID       string              `json:"id"`
	Status   CodeExecutionStatus `json:"status"`
	Output   string              `json:"output"`
	WallTime float32             `json:"wall_time"`
	Memory   int32               `json:"memory"`
}
