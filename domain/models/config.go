package models

type RunnerConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RunnerConfigDetail struct {
	RunnerConfig
	BuildScript string `json:"build_script"`
	RunScript   string `json:"run_script"`
}

type CompareConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
