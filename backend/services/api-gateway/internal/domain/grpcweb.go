package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type GRPCWebAuthMode string

const (
	GRPCWebAuthPublic   GRPCWebAuthMode = "public"
	GRPCWebAuthOptional GRPCWebAuthMode = "optional"
	GRPCWebAuthRequired GRPCWebAuthMode = "required"
)

type GRPCWebMethodPolicy struct {
	FullMethod string
	Downstream string
	AuthMode   GRPCWebAuthMode
	Roles      []string
	Timeout    time.Duration
}

func (p GRPCWebMethodPolicy) Validate() error {
	var errs []error
	service, method, ok := ParseFullGRPCMethod(p.FullMethod)
	if !ok {
		errs = append(errs, errors.New("full_method must use /package.Service/Method format"))
	}
	if strings.ContainsAny(p.FullMethod, " \t\r\n") {
		errs = append(errs, errors.New("full_method must not contain whitespace"))
	}
	if strings.TrimSpace(service) == "" || strings.TrimSpace(method) == "" {
		errs = append(errs, errors.New("full_method must include a service and method"))
	}
	if strings.TrimSpace(p.Downstream) == "" {
		errs = append(errs, errors.New("downstream is required"))
	} else if strings.ContainsAny(p.Downstream, " \t\r\n") {
		errs = append(errs, errors.New("downstream must not contain whitespace"))
	}
	switch p.AuthMode {
	case GRPCWebAuthPublic, GRPCWebAuthOptional, GRPCWebAuthRequired:
	default:
		errs = append(errs, fmt.Errorf("unsupported auth mode %q", p.AuthMode))
	}
	if p.AuthMode != GRPCWebAuthRequired && len(p.Roles) > 0 {
		errs = append(errs, errors.New("roles require auth mode required"))
	}
	if p.Timeout <= 0 {
		errs = append(errs, errors.New("timeout must be positive"))
	}
	for _, role := range p.Roles {
		if strings.TrimSpace(role) == "" {
			errs = append(errs, errors.New("roles must not contain empty values"))
			break
		}
	}
	return errors.Join(errs...)
}

func (p GRPCWebMethodPolicy) Service() string {
	service, _, _ := ParseFullGRPCMethod(p.FullMethod)
	return service
}

func (p GRPCWebMethodPolicy) RequiresToken() bool {
	return p.AuthMode == GRPCWebAuthOptional || p.AuthMode == GRPCWebAuthRequired
}

func ParseFullGRPCMethod(fullMethod string) (service string, method string, ok bool) {
	fullMethod = strings.TrimSpace(fullMethod)
	if !strings.HasPrefix(fullMethod, "/") {
		return "", "", false
	}
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}
