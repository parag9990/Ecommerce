package clients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/observability"

	"google.golang.org/grpc"
)

type Clients struct {
	Auth         AuthServiceClient
	User         UserServiceClient
	Product      ProductServiceClient
	Cart         CartServiceClient
	Wishlist     WishlistServiceClient
	Order        OrderServiceClient
	Payment      PaymentServiceClient
	Search       SearchServiceClient
	CMS          CMSServiceClient
	Session      SessionServiceClient
	Notification NotificationServiceClient
	Superadmin   SuperadminServiceClient

	conns       map[Downstream]*grpc.ClientConn
	descriptors map[Downstream]ServiceDescriptor
	closeMu     sync.Mutex
	closed      bool
	logger      *slog.Logger
}

type Option func(*clientOptions)

type clientOptions struct {
	metrics *observability.Metrics
}

func WithMetrics(metrics *observability.Metrics) Option {
	return func(opts *clientOptions) {
		opts.metrics = metrics
	}
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger, options ...Option) (*Clients, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	clientOpts := clientOptions{}
	for _, option := range options {
		if option != nil {
			option(&clientOpts)
		}
	}
	observabilityConfig := cfg.Observability.Normalize(cfg.ServiceName, cfg.Environment)
	dialer := NewGRPCDialerWithProviders(DialOptions{
		TLSEnabled:       cfg.GRPCTLSEnabled,
		DialTimeout:      cfg.GRPCDialTimeout,
		AllowUnavailable: cfg.GRPCAllowUnavailableDownstreams,
	}, logger, []DialOptionProvider{
		func(descriptor ServiceDescriptor) grpc.DialOption {
			return grpc.WithChainUnaryInterceptor(UnaryDeadlineInterceptor(descriptor.DefaultTimeout))
		},
		func(descriptor ServiceDescriptor) grpc.DialOption {
			return grpc.WithChainStreamInterceptor(StreamDeadlineInterceptor(descriptor.DefaultTimeout))
		},
		func(descriptor ServiceDescriptor) grpc.DialOption {
			return grpc.WithChainUnaryInterceptor(observability.UnaryClientInterceptor(
				observabilityConfig,
				clientOpts.metrics,
				string(descriptor.Name),
			))
		},
		func(descriptor ServiceDescriptor) grpc.DialOption {
			return grpc.WithChainStreamInterceptor(observability.StreamClientInterceptor(
				observabilityConfig,
				clientOpts.metrics,
				string(descriptor.Name),
			))
		},
	})
	return NewWithDialer(ctx, ServiceDescriptorsFromConfig(cfg), dialer, logger)
}

func NewWithDialer(ctx context.Context, descriptors []ServiceDescriptor, dialer Dialer, logger *slog.Logger) (*Clients, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if dialer == nil {
		return nil, errors.New("grpc dialer is required")
	}
	if err := validateServiceDescriptors(descriptors); err != nil {
		return nil, err
	}

	registry := &Clients{
		conns:       make(map[Downstream]*grpc.ClientConn, len(descriptors)),
		descriptors: make(map[Downstream]ServiceDescriptor, len(descriptors)),
		logger:      logger,
	}
	for _, descriptor := range descriptors {
		conn, err := dialer.Dial(ctx, descriptor)
		if err != nil {
			_ = registry.Close()
			return nil, fmt.Errorf("dial %s service: %w", descriptor.Name, err)
		}
		if conn == nil {
			_ = registry.Close()
			return nil, fmt.Errorf("dial %s service: nil connection", descriptor.Name)
		}
		if err := registry.attach(descriptor, conn); err != nil {
			_ = conn.Close()
			_ = registry.Close()
			return nil, err
		}
	}
	if err := registry.validateComplete(); err != nil {
		_ = registry.Close()
		return nil, err
	}
	return registry, nil
}

func (c *Clients) Connection(service Downstream) (*grpc.ClientConn, bool) {
	if c == nil {
		return nil, false
	}
	conn, ok := c.conns[service]
	return conn, ok
}

func (c *Clients) Client(service Downstream) (OutboundClient, bool) {
	if c == nil {
		return nil, false
	}
	switch service {
	case DownstreamAuth:
		return c.Auth, c.Auth != nil
	case DownstreamUser:
		return c.User, c.User != nil
	case DownstreamProduct:
		return c.Product, c.Product != nil
	case DownstreamCart:
		return c.Cart, c.Cart != nil
	case DownstreamWishlist:
		return c.Wishlist, c.Wishlist != nil
	case DownstreamOrder:
		return c.Order, c.Order != nil
	case DownstreamPayment:
		return c.Payment, c.Payment != nil
	case DownstreamSearch:
		return c.Search, c.Search != nil
	case DownstreamCMS:
		return c.CMS, c.CMS != nil
	case DownstreamSession:
		return c.Session, c.Session != nil
	case DownstreamNotification:
		return c.Notification, c.Notification != nil
	case DownstreamSuperadmin:
		return c.Superadmin, c.Superadmin != nil
	default:
		return nil, false
	}
}

func (c *Clients) Close() error {
	if c == nil {
		return nil
	}
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true

	var errs []error
	for _, service := range serviceOrder {
		conn := c.conns[service]
		if conn == nil {
			continue
		}
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s service connection: %w", service, err))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.InfoContext(context.Background(), "grpc_clients_closed")
	}
	return nil
}

func (c *Clients) attach(descriptor ServiceDescriptor, conn *grpc.ClientConn) error {
	if _, exists := c.conns[descriptor.Name]; exists {
		return fmt.Errorf("duplicate grpc client for %s", descriptor.Name)
	}
	c.conns[descriptor.Name] = conn
	c.descriptors[descriptor.Name] = descriptor

	switch descriptor.Name {
	case DownstreamAuth:
		c.Auth = newAuthServiceClient(descriptor, conn)
	case DownstreamUser:
		c.User = newUserServiceClient(descriptor, conn)
	case DownstreamProduct:
		c.Product = newProductServiceClient(descriptor, conn)
	case DownstreamCart:
		c.Cart = newCartServiceClient(descriptor, conn)
	case DownstreamWishlist:
		c.Wishlist = newWishlistServiceClient(descriptor, conn)
	case DownstreamOrder:
		c.Order = newOrderServiceClient(descriptor, conn)
	case DownstreamPayment:
		c.Payment = newPaymentServiceClient(descriptor, conn)
	case DownstreamSearch:
		c.Search = newSearchServiceClient(descriptor, conn)
	case DownstreamCMS:
		c.CMS = newCMSServiceClient(descriptor, conn)
	case DownstreamSession:
		c.Session = newSessionServiceClient(descriptor, conn)
	case DownstreamNotification:
		c.Notification = newNotificationServiceClient(descriptor, conn)
	case DownstreamSuperadmin:
		c.Superadmin = newSuperadminServiceClient(descriptor, conn)
	default:
		return fmt.Errorf("unsupported downstream service %q", descriptor.Name)
	}
	return nil
}

func (c *Clients) validateComplete() error {
	var errs []error
	for _, service := range serviceOrder {
		if _, ok := c.descriptors[service]; !ok {
			errs = append(errs, fmt.Errorf("%s grpc client is required", service))
			continue
		}
		if _, ok := c.Client(service); !ok {
			errs = append(errs, fmt.Errorf("%s grpc client is not initialized", service))
		}
	}
	return errors.Join(errs...)
}

func validateServiceDescriptors(descriptors []ServiceDescriptor) error {
	expected := make(map[Downstream]struct{}, len(serviceOrder))
	for _, service := range serviceOrder {
		expected[service] = struct{}{}
	}

	seen := make(map[Downstream]struct{}, len(descriptors))
	var errs []error
	for _, descriptor := range descriptors {
		if err := descriptor.Validate(); err != nil {
			errs = append(errs, err)
		}
		if descriptor.Name == "" {
			continue
		}
		if _, ok := expected[descriptor.Name]; !ok {
			errs = append(errs, fmt.Errorf("unsupported downstream service %q", descriptor.Name))
			continue
		}
		if _, duplicate := seen[descriptor.Name]; duplicate {
			errs = append(errs, fmt.Errorf("duplicate grpc client descriptor for %s", descriptor.Name))
			continue
		}
		seen[descriptor.Name] = struct{}{}
	}
	for _, service := range serviceOrder {
		if _, ok := seen[service]; !ok {
			errs = append(errs, fmt.Errorf("%s grpc client descriptor is required", service))
		}
	}
	return errors.Join(errs...)
}
