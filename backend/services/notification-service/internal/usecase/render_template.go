package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/domain"
	"github.com/example/ecommerce-platform/backend/services/notification-service/internal/repository"
)

// Renderer produces provider-independent content from a stored template.
type Renderer interface {
	Render(ctx context.Context, req domain.RenderRequest) (domain.RenderedMessage, error)
}

type TemplateRenderer struct {
	templates repository.TemplateRepository
}

var _ Renderer = (*TemplateRenderer)(nil)

func NewTemplateRenderer(templates repository.TemplateRepository) (*TemplateRenderer, error) {
	if nilDependency(templates) {
		return nil, errors.New("template repository is required")
	}
	return &TemplateRenderer{templates: templates}, nil
}

func (r *TemplateRenderer) Render(
	ctx context.Context,
	req domain.RenderRequest,
) (domain.RenderedMessage, error) {
	if r == nil || nilDependency(r.templates) {
		return domain.RenderedMessage{}, errors.New("template renderer is not configured")
	}
	if ctx == nil {
		return domain.RenderedMessage{}, fmt.Errorf("%w: context is required", domain.ErrInvalidRenderRequest)
	}
	if err := ctx.Err(); err != nil {
		return domain.RenderedMessage{}, err
	}

	variables, err := approvedVariables(req)
	if err != nil {
		return domain.RenderedMessage{}, err
	}

	stored, err := r.templates.FindLatestActive(ctx, string(req.TemplateKey), req.Channel)
	if err != nil {
		return domain.RenderedMessage{}, fmt.Errorf(
			"load active template %q for channel %q: %w", req.TemplateKey, req.Channel, err)
	}
	if err := validateSelectedTemplate(stored, req); err != nil {
		return domain.RenderedMessage{}, err
	}

	subject, err := renderText("subject", stored.Subject, variables)
	if err != nil {
		return domain.RenderedMessage{}, err
	}
	if req.Channel == domain.ChannelEmail && strings.ContainsAny(subject, "\r\n") {
		return domain.RenderedMessage{}, fmt.Errorf("%w: email subject contains a line break", domain.ErrTemplateRender)
	}

	body, err := renderText("body", stored.Body, variables)
	if err != nil {
		return domain.RenderedMessage{}, err
	}

	return domain.RenderedMessage{
		TemplateID:      stored.ID,
		TemplateKey:     req.TemplateKey,
		TemplateVersion: stored.Version,
		Channel:         req.Channel,
		Subject:         subject,
		Body:            body,
	}, nil
}

func validateSelectedTemplate(stored domain.Template, req domain.RenderRequest) error {
	if err := stored.Validate(); err != nil {
		return fmt.Errorf("%w: stored template is invalid: %w", domain.ErrTemplateRender, err)
	}
	if stored.TemplateKey != string(req.TemplateKey) ||
		stored.Channel != req.Channel ||
		stored.Status != domain.TemplateStatusActive {
		return fmt.Errorf("%w: repository returned a template outside the requested active revision",
			domain.ErrTemplateRender)
	}
	return nil
}

func renderText(part, source string, variables map[string]string) (string, error) {
	if source == "" {
		return "", nil
	}
	parsed, err := template.New(part).Option("missingkey=error").Parse(source)
	if err != nil {
		return "", fmt.Errorf("%w: parse %s: %v", domain.ErrTemplateRender, part, err)
	}

	var output bytes.Buffer
	if err := parsed.Execute(&output, variables); err != nil {
		return "", fmt.Errorf("%w: render %s: %v", domain.ErrTemplateRender, part, err)
	}
	return output.String(), nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
