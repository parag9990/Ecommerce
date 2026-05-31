package validation

import (
	"fmt"
	"strings"
)

func RequiredText(field string, value string, maxLen int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if maxLen > 0 && len(value) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func OptionalText(field string, value string, maxLen int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if maxLen > 0 && len(value) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", field, maxLen)
	}
	return nil
}

func SplitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			out = append(out, cleaned)
		}
	}
	return out
}
