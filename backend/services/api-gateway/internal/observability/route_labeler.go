package observability

import (
	"net/http"
	"sort"
	"strings"

	"ecommerce/api-gateway/internal/domain"
)

const (
	RouteUnknown   = "unknown"
	RouteUnmatched = "unmatched"
)

type RouteLabeler interface {
	RouteTemplate(method, path string) string
}

type RoutePattern struct {
	Method   string
	Template string
}

type StaticRoute struct {
	Method string
	Path   string
}

type routeLabeler struct {
	patterns []compiledPattern
}

type compiledPattern struct {
	method       string
	template     string
	segments     []compiledSegment
	staticWeight int
}

type compiledSegment struct {
	value string
	param bool
}

func NewRouteLabeler(routes []domain.RouteDefinition, staticRoutes ...StaticRoute) RouteLabeler {
	patterns := make([]compiledPattern, 0, len(routes)+len(staticRoutes))
	for _, route := range routes {
		if compiled, ok := compilePattern(string(route.Method), route.Path); ok {
			patterns = append(patterns, compiled)
		}
	}
	for _, route := range staticRoutes {
		if compiled, ok := compilePattern(route.Method, route.Path); ok {
			patterns = append(patterns, compiled)
		}
	}
	sort.SliceStable(patterns, func(i, j int) bool {
		if len(patterns[i].segments) != len(patterns[j].segments) {
			return len(patterns[i].segments) > len(patterns[j].segments)
		}
		return patterns[i].staticWeight > patterns[j].staticWeight
	})
	return routeLabeler{patterns: patterns}
}

func (l routeLabeler) RouteTemplate(method, path string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	path = cleanPath(path)
	for _, pattern := range l.patterns {
		if pattern.method != "" && pattern.method != method {
			continue
		}
		if pattern.match(path) {
			return pattern.template
		}
	}
	return RouteUnmatched
}

func RouteTemplateFromRequest(r *http.Request, labeler RouteLabeler) string {
	if r == nil {
		return RouteUnknown
	}
	if labeler != nil {
		if route := labeler.RouteTemplate(r.Method, r.URL.Path); route != "" && route != RouteUnmatched {
			return route
		}
	}
	if pattern := pathFromServeMuxPattern(r.Pattern); pattern != "" {
		return pattern
	}
	if labeler != nil {
		return RouteUnmatched
	}
	return RouteUnknown
}

func (p compiledPattern) match(path string) bool {
	segments := splitPath(path)
	if len(segments) != len(p.segments) {
		return false
	}
	for i, segment := range segments {
		expected := p.segments[i]
		if expected.param {
			if segment == "" {
				return false
			}
			continue
		}
		if expected.value != segment {
			return false
		}
	}
	return true
}

func compilePattern(method, template string) (compiledPattern, bool) {
	template = cleanPath(template)
	if template == "" {
		return compiledPattern{}, false
	}
	segments := splitPath(template)
	compiled := compiledPattern{
		method:   strings.ToUpper(strings.TrimSpace(method)),
		template: template,
		segments: make([]compiledSegment, 0, len(segments)),
	}
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			compiled.segments = append(compiled.segments, compiledSegment{param: true})
			continue
		}
		compiled.staticWeight++
		compiled.segments = append(compiled.segments, compiledSegment{value: segment})
	}
	return compiled, true
}

func cleanPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path
}

func splitPath(path string) []string {
	path = strings.Trim(cleanPath(path), "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func pathFromServeMuxPattern(pattern string) string {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return ""
	}
	parts := strings.Fields(pattern)
	if len(parts) == 0 {
		return ""
	}
	path := parts[len(parts)-1]
	if strings.HasPrefix(path, "/") {
		return cleanPath(path)
	}
	return ""
}
