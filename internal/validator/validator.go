package validator

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AppValidator interface {
	Validate(data any) []ErrorResponse
}

type appValidator struct {
	validate *validator.Validate
}

func isUUIDSlice(fl validator.FieldLevel) bool {
	slice, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}

	for _, s := range slice {
		if _, err := uuid.Parse(s); err != nil {
			return false
		}
	}

	return true
}

func NewAppValidator() AppValidator {
	validate := validator.New()

	validate.RegisterValidation("uuid_slice", isUUIDSlice)

	return &appValidator{validate: validate}
}

type ErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v *appValidator) Validate(data any) []ErrorResponse {
	errors := []ErrorResponse{}

	if errs := v.validate.Struct(data); errs != nil {

		for _, err := range errs.(validator.ValidationErrors) {
			fieldName := getJSONFieldName(data, err.Field())
			errors = append(errors, ErrorResponse{
				Field:   fieldName,
				Message: err.Tag(),
			})
		}
	}

	return errors
}

// Thanks ChatGPT
func getJSONFieldName(obj any, fieldName string) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference if it's a pointer
	}
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName // Fallback to struct field name
	}
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return fieldName // Fallback if no JSON tag or it's ignored
	}
	return jsonTag
}
