package clients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const defaultDialTimeout = 3 * time.Second

type Dialer interface {
	Dial(ctx context.Context, descriptor ServiceDescriptor) (*grpc.ClientConn, error)
}

type DialOptions struct {
	TLSEnabled       bool
	DialTimeout      time.Duration
	AllowUnavailable bool
}

type DialOptionProvider func(ServiceDescriptor) grpc.DialOption

type grpcDialer struct {
	options   DialOptions
	logger    *slog.Logger
	extra     []grpc.DialOption
	providers []DialOptionProvider
}

func NewGRPCDialer(options DialOptions, logger *slog.Logger, extra ...grpc.DialOption) Dialer {
	if options.DialTimeout <= 0 {
		options.DialTimeout = defaultDialTimeout
	}
	copiedExtra := make([]grpc.DialOption, len(extra))
	copy(copiedExtra, extra)
	return &grpcDialer{
		options: options,
		logger:  logger,
		extra:   copiedExtra,
	}
}

func NewGRPCDialerWithProviders(options DialOptions, logger *slog.Logger, providers []DialOptionProvider, extra ...grpc.DialOption) Dialer {
	dialer := NewGRPCDialer(options, logger, extra...)
	grpcDialer, ok := dialer.(*grpcDialer)
	if !ok {
		return dialer
	}
	grpcDialer.providers = append(grpcDialer.providers, providers...)
	return grpcDialer
}

func (d *grpcDialer) Dial(ctx context.Context, descriptor ServiceDescriptor) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := descriptor.Validate(); err != nil {
		return nil, err
	}

	var transport credentials.TransportCredentials
	if d.options.TLSEnabled {
		transport = credentials.NewClientTLSFromCert(nil, "")
	} else {
		transport = insecure.NewCredentials()
	}

	dialCtx, cancel := context.WithTimeout(ctx, d.options.DialTimeout)
	defer cancel()

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(transport),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  100 * time.Millisecond,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   3 * time.Second,
			},
			MinConnectTimeout: d.options.DialTimeout,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	dialOptions = append(dialOptions, d.extra...)
	for _, provider := range d.providers {
		if provider == nil {
			continue
		}
		option := provider(descriptor)
		if option != nil {
			dialOptions = append(dialOptions, option)
		}
	}

	conn, err := grpc.NewClient(descriptor.Target, dialOptions...)
	if err != nil {
		return nil, err
	}
	conn.Connect()
	if d.options.AllowUnavailable {
		if d.logger != nil {
			d.logger.WarnContext(ctx, "grpc_client_started_without_readiness_wait",
				"service", descriptor.Name,
				"target", descriptor.Target,
			)
		}
		return conn, nil
	}

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			if d.logger != nil {
				d.logger.InfoContext(ctx, "grpc_client_connected",
					"service", descriptor.Name,
					"target", descriptor.Target,
				)
			}
			return conn, nil
		}

		if !conn.WaitForStateChange(dialCtx, state) {
			closeErr := conn.Close()
			if err := dialCtx.Err(); err != nil {
				if closeErr != nil {
					return nil, errors.Join(err, closeErr)
				}
				return nil, err
			}
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, fmt.Errorf("%s connection did not become ready", descriptor.Name)
		}
	}
}
