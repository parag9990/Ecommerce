package domain

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ContractValidationError struct {
	Errors []error
}

func (e ContractValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "contract validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("contract validation failed with %d errors: %s", len(e.Errors), e.Errors[0].Error())
}
