package requests

import (
	configPB "github.com/CSKU-Lab/main-server/genproto/config/v1"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateRunnerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRunnerRequest struct {
	Name         *string       `json:"name"`
	Description  *string       `json:"description"`
	BuildScript  *string       `json:"build_script"`
	RunScript    *string       `json:"run_script"`
	InitialFiles *[]ConfigFile `json:"initial_files"`
}

func (cr *CreateRunnerRequest) Validate() error {
	return validation.ValidateStruct(cr,
		validation.Field(&cr.Name, validation.Required),
		validation.Field(&cr.Description, validation.NilOrNotEmpty),
	)
}

func (ur *UpdateRunnerRequest) Validate() error {
	return nil
}

type TestRunnerRequest struct {
	InitialFiles []ConfigFile `json:"initial_files"`
	Input        string       `json:"input"`
	RunScript    string       `json:"run_script"`
	BuildScript  string       `json:"build_script"`
}

type ConfigFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type CreateCompareRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	BuildScript string       `json:"build_script"`
	RunScript   string       `json:"run_script"`
	RunName     string       `json:"run_name"`
	Files       []ConfigFile `json:"files"`
}

type UpdateCompareRequest struct {
	Name        *string      `json:"name"`
	Description *string      `json:"description"`
	BuildScript *string      `json:"build_script"`
	RunScript   *string      `json:"run_script"`
	RunName     *string      `json:"run_name"`
	Files       []ConfigFile `json:"files"`
}

func (cc *CreateCompareRequest) Validate() error {
	return validation.ValidateStruct(cc,
		validation.Field(&cc.Name, validation.Required),
		validation.Field(&cc.Description, validation.Required),
		validation.Field(&cc.BuildScript, validation.Required),
		validation.Field(&cc.RunScript, validation.Required),
		validation.Field(&cc.RunName, validation.Required),
		validation.Field(&cc.Files, validation.NilOrNotEmpty),
	)
}

func (uc *UpdateCompareRequest) Validate() error {
	return validation.ValidateStruct(uc,
		validation.Field(&uc.Name, validation.NilOrNotEmpty),
		validation.Field(&uc.Description, validation.NilOrNotEmpty),
		validation.Field(&uc.BuildScript, validation.NilOrNotEmpty),
		validation.Field(&uc.RunScript, validation.NilOrNotEmpty),
		validation.Field(&uc.RunName, validation.NilOrNotEmpty),
		validation.Field(&uc.Files, validation.NilOrNotEmpty),
	)
}

func MapConfigFilesToPB(files []ConfigFile) []*configPB.File {
	pbFiles := make([]*configPB.File, len(files))
	for i, file := range files {
		pbFiles[i] = &configPB.File{
			Name:    file.Name,
			Content: file.Content,
		}
	}
	return pbFiles
}
