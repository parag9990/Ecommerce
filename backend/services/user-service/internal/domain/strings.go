package domain

import "strings"

func cleanOptionalUpper(value *string) *string {
	cleaned := cleanOptional(value)
	if cleaned == nil {
		return nil
	}

	upper := strings.ToUpper(*cleaned)
	return &upper
}
