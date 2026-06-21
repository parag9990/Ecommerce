package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	Lookup func(string) (string, bool)
}

func System() Env { return Env{Lookup: os.LookupEnv} }

func (e Env) String(name, fallback string) string {
	if value, ok := e.lookup(name); ok && value != "" {
		return value
	}
	return fallback
}

func (e Env) Required(name string) (string, error) {
	value := e.String(name, "")
	if value == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func (e Env) Bool(name string, fallback bool) (bool, error) {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
	}
	return parsed, nil
}

func (e Env) Int(name string, fallback, minimum int) (int, error) {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		value = strconv.Itoa(fallback)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum {
		return 0, fmt.Errorf("%s must be an integer greater than or equal to %d", name, minimum)
	}
	return parsed, nil
}

func (e Env) Float64(name string, fallback, minimum, maximum float64) (float64, error) {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		value = strconv.FormatFloat(fallback, 'f', -1, 64)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < minimum || parsed > maximum {
		return 0, fmt.Errorf("%s must be between %g and %g", name, minimum, maximum)
	}
	return parsed, nil
}

func (e Env) CSV(name string, fallback []string) []string {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}

func (e Env) Duration(name string, fallback time.Duration) (time.Duration, error) {
	value, ok := e.lookup(name)
	if !ok || value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", name)
	}
	return parsed, nil
}

func (e Env) lookup(name string) (string, bool) {
	if e.Lookup == nil {
		return "", false
	}
	value, ok := e.Lookup(name)
	return strings.TrimSpace(value), ok
}

func JoinErrors(errs ...error) error {
	var present []error
	for _, err := range errs {
		if err != nil {
			present = append(present, err)
		}
	}
	return errors.Join(present...)
}
