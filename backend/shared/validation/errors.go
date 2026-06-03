package validation

const CodeValidationError = "VALIDATION_ERROR"

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Error struct {
	Fields []FieldError
}

func (e Error) Error() string {
	return "validation failed"
}

func (e Error) HasErrors() bool {
	return len(e.Fields) > 0
}

func (e *Error) Add(field string, code string, message string) {
	e.Fields = append(e.Fields, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
}

func (e *Error) AddRequired(field string) {
	e.Add(field, "required", field+" is required")
}

func SingleFieldError(field string, code string, message string) Error {
	var err Error
	err.Add(field, code, message)
	return err
}
