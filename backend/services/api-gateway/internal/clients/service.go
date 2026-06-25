package clients

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"ecommerce/api-gateway/internal/config"
)

type Downstream string

const (
	DownstreamAuth         Downstream = "auth"
	DownstreamUser         Downstream = "user"
	DownstreamProduct      Downstream = "product"
	DownstreamCart         Downstream = "cart"
	DownstreamWishlist     Downstream = "wishlist"
	DownstreamOrder        Downstream = "order"
	DownstreamPayment      Downstream = "payment"
	DownstreamSearch       Downstream = "search"
	DownstreamCMS          Downstream = "cms"
	DownstreamSession      Downstream = "session"
	DownstreamNotification Downstream = "notification"
	DownstreamSuperadmin   Downstream = "superadmin"
)

var serviceOrder = []Downstream{
	DownstreamAuth,
	DownstreamUser,
	DownstreamProduct,
	DownstreamCart,
	DownstreamWishlist,
	DownstreamOrder,
	DownstreamPayment,
	DownstreamSearch,
	DownstreamCMS,
	DownstreamSession,
	DownstreamNotification,
	DownstreamSuperadmin,
}

type ServiceDescriptor struct {
	Name           Downstream
	Target         string
	HealthService  string
	DefaultTimeout time.Duration
}

func (d ServiceDescriptor) Validate() error {
	var errs []error
	if d.Name == "" {
		errs = append(errs, errors.New("service name is required"))
	}
	if strings.TrimSpace(d.Target) == "" {
		errs = append(errs, fmt.Errorf("%s target is required", d.Name))
	} else if strings.ContainsAny(d.Target, " \t\r\n") {
		errs = append(errs, fmt.Errorf("%s target must not contain whitespace", d.Name))
	}
	if strings.TrimSpace(d.HealthService) == "" {
		errs = append(errs, fmt.Errorf("%s health service is required", d.Name))
	}
	if d.DefaultTimeout <= 0 {
		errs = append(errs, fmt.Errorf("%s default timeout must be positive", d.Name))
	}
	return errors.Join(errs...)
}

func ServiceDescriptorsFromConfig(cfg config.Config) []ServiceDescriptor {
	return []ServiceDescriptor{
		newServiceDescriptor(DownstreamAuth, cfg.AuthGRPCAddr, "ecommerce.auth.v1.AuthService"),
		newServiceDescriptor(DownstreamUser, cfg.UserGRPCAddr, "ecommerce.user.v1.UserService"),
		newServiceDescriptor(DownstreamProduct, cfg.ProductGRPCAddr, "ecommerce.product.v1.ProductService"),
		newServiceDescriptor(DownstreamCart, cfg.CartGRPCAddr, "ecommerce.cart.v1.CartService"),
		newServiceDescriptor(DownstreamWishlist, cfg.WishlistGRPCAddr, "ecommerce.wishlist.v1.WishlistService"),
		newServiceDescriptor(DownstreamOrder, cfg.OrderGRPCAddr, "ecommerce.order.v1.OrderService"),
		newServiceDescriptor(DownstreamPayment, cfg.PaymentGRPCAddr, "ecommerce.payment.v1.PaymentService"),
		newServiceDescriptor(DownstreamSearch, cfg.SearchGRPCAddr, "ecommerce.search.v1.SearchService"),
		newServiceDescriptor(DownstreamCMS, cfg.CMSGRPCAddr, "ecommerce.cms.v1.CMSService"),
		newServiceDescriptor(DownstreamSession, cfg.SessionGRPCAddr, "ecommerce.session.v1.SessionService"),
		newServiceDescriptor(DownstreamNotification, cfg.NotificationGRPCAddr, "ecommerce.notification.v1.NotificationService"),
		newServiceDescriptor(DownstreamSuperadmin, cfg.SuperadminGRPCAddr, "ecommerce.superadmin.v1.SuperadminService"),
	}
}

func newServiceDescriptor(name Downstream, target string, healthService string) ServiceDescriptor {
	return ServiceDescriptor{
		Name:           name,
		Target:         strings.TrimSpace(target),
		HealthService:  healthService,
		DefaultTimeout: TimeoutFor(name),
	}
}
