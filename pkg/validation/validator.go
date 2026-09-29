package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (e *ValidationErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return strings.Join(msgs, "; ")
}

func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errs []ValidationError
	for _, fe := range err.(validator.ValidationErrors) {
		errs = append(errs, ValidationError{
			Field:   strings.ToLower(fe.Field()),
			Message: fieldMessage(fe),
		})
	}
	return &ValidationErrors{Errors: errs}
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "email":
		return "must be a valid email address"
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "latitude":
		return "must be a valid latitude (-90 to 90)"
	case "longitude":
		return "must be a valid longitude (-180 to 180)"
	default:
		return fmt.Sprintf("failed validation: %s", fe.Tag())
	}
}
