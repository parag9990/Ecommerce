package ratelimit

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Dimension string

const (
	DimensionIP     Dimension = "ip"
	DimensionUser   Dimension = "user"
	DimensionRoute  Dimension = "route"
	DimensionTarget Dimension = "target"
)

type Policy struct {
	Name         string
	Dimension    Dimension
	RouteIDs     []string
	Methods      []string
	PathPrefixes []string
	Limit        int64
	Window       time.Duration
	Cost         int64
	TargetFields []string
}

func DefaultPolicies(defaultIPLimit int64, defaultIPWindow time.Duration) []Policy {
	if defaultIPLimit <= 0 {
		defaultIPLimit = 600
	}
	if defaultIPWindow <= 0 {
		defaultIPWindow = time.Minute
	}

	return []Policy{
		{
			Name:      "global.ip",
			Dimension: DimensionIP,
			Limit:     defaultIPLimit,
			Window:    defaultIPWindow,
			Cost:      1,
		},
		{
			Name:      "auth.login.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"auth.login"},
			Limit:     5,
			Window:    10 * time.Minute,
			Cost:      1,
		},
		{
			Name:      "auth.otp_send.target",
			Dimension: DimensionTarget,
			RouteIDs:  []string{"auth.otp_send"},
			Limit:     3,
			Window:    15 * time.Minute,
			Cost:      1,
			TargetFields: []string{
				"target",
			},
		},
		{
			Name:      "auth.otp_send.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"auth.otp_send"},
			Limit:     10,
			Window:    15 * time.Minute,
			Cost:      1,
		},
		{
			Name:      "auth.password_reset.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"auth.password_forgot", "auth.password_reset"},
			Limit:     5,
			Window:    15 * time.Minute,
			Cost:      1,
		},
		{
			Name:      "search.public.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"search.products", "search.autocomplete"},
			Limit:     120,
			Window:    time.Minute,
			Cost:      1,
		},
		{
			Name:      "catalog.public.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"product.list", "product.detail", "category.list"},
			Limit:     240,
			Window:    time.Minute,
			Cost:      1,
		},
		{
			Name:      "cart.user",
			Dimension: DimensionUser,
			RouteIDs:  []string{"cart.add_item", "cart.update_item", "cart.remove_item", "cart.merge", "cart.coupon_preview"},
			Limit:     60,
			Window:    time.Minute,
			Cost:      1,
		},
		{
			Name:      "wishlist.user",
			Dimension: DimensionUser,
			RouteIDs:  []string{"wishlist.add", "wishlist.remove", "wishlist.move_to_cart"},
			Limit:     60,
			Window:    time.Minute,
			Cost:      1,
		},
		{
			Name:      "checkout.user",
			Dimension: DimensionUser,
			RouteIDs:  []string{"order.checkout"},
			Limit:     10,
			Window:    10 * time.Minute,
			Cost:      1,
		},
		{
			Name:      "seller.mutation.user",
			Dimension: DimensionUser,
			RouteIDs: []string{
				"seller.profile_update",
				"seller.product_create",
				"seller.product_update",
				"seller.product_publish",
				"seller.order.fulfillment",
				"cms.coupon_create",
				"cms.coupon_update",
				"cms.campaign_create",
			},
			Limit:  60,
			Window: time.Minute,
			Cost:   1,
		},
		{
			Name:      "admin.mutation.user",
			Dimension: DimensionUser,
			RouteIDs: []string{
				"payment.refund",
				"admin.user_status",
				"admin.seller_status",
				"admin.refund_review",
				"admin.setting_update",
			},
			Limit:  30,
			Window: time.Minute,
			Cost:   1,
		},
		{
			Name:      "webhook.provider.ip",
			Dimension: DimensionIP,
			RouteIDs:  []string{"payment.webhook"},
			Limit:     120,
			Window:    time.Minute,
			Cost:      1,
		},
	}
}

func MatchPolicies(policies []Policy, routeID string, method string, path string, dimensions ...Dimension) []Policy {
	allowedDimensions := make(map[Dimension]struct{}, len(dimensions))
	for _, dimension := range dimensions {
		allowedDimensions[dimension] = struct{}{}
	}

	matches := make([]Policy, 0, len(policies))
	for _, policy := range policies {
		if len(allowedDimensions) > 0 {
			if _, ok := allowedDimensions[policy.Dimension]; !ok {
				continue
			}
		}
		if policy.matches(routeID, method, path) {
			matches = append(matches, policy)
		}
	}
	return matches
}

func ValidatePolicies(policies []Policy) error {
	var errs []error
	seen := make(map[string]struct{}, len(policies))
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			errs = append(errs, err)
		}
		if _, ok := seen[policy.Name]; ok {
			errs = append(errs, fmt.Errorf("rate limit policy %q is duplicated", policy.Name))
		}
		seen[policy.Name] = struct{}{}
	}
	return errors.Join(errs...)
}

func (p Policy) Validate() error {
	var errs []error
	if strings.TrimSpace(p.Name) == "" {
		errs = append(errs, errors.New("rate limit policy name is required"))
	}
	if !isSupportedDimension(p.Dimension) {
		errs = append(errs, fmt.Errorf("rate limit policy %q has unsupported dimension %q", p.Name, p.Dimension))
	}
	if p.Limit <= 0 {
		errs = append(errs, fmt.Errorf("rate limit policy %q limit must be positive", p.Name))
	}
	if p.Window <= 0 {
		errs = append(errs, fmt.Errorf("rate limit policy %q window must be positive", p.Name))
	}
	if p.Cost < 0 {
		errs = append(errs, fmt.Errorf("rate limit policy %q cost must not be negative", p.Name))
	}
	if p.Dimension == DimensionTarget && len(p.TargetFields) == 0 {
		errs = append(errs, fmt.Errorf("rate limit policy %q target fields are required", p.Name))
	}
	for _, method := range p.Methods {
		switch strings.ToUpper(strings.TrimSpace(method)) {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodPut:
		default:
			errs = append(errs, fmt.Errorf("rate limit policy %q has unsupported method %q", p.Name, method))
		}
	}
	return errors.Join(errs...)
}

func (p Policy) RequestCost() int64 {
	if p.Cost <= 0 {
		return 1
	}
	return p.Cost
}

func (p Policy) matches(routeID string, method string, path string) bool {
	if len(p.RouteIDs) > 0 && !containsFold(p.RouteIDs, routeID) {
		return false
	}
	if len(p.Methods) > 0 && !containsFold(p.Methods, method) {
		return false
	}
	if len(p.PathPrefixes) > 0 && !hasAnyPathPrefix(path, p.PathPrefixes) {
		return false
	}
	return true
}

func isSupportedDimension(dimension Dimension) bool {
	switch dimension {
	case DimensionIP, DimensionUser, DimensionRoute, DimensionTarget:
		return true
	default:
		return false
	}
}

func containsFold(values []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}
	return false
}

func hasAnyPathPrefix(path string, prefixes []string) bool {
	for _, prefix := range prefixes {
		prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
		if prefix == "" {
			continue
		}
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}
