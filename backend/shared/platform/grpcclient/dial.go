package grpcclient

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Target, ServerName string
	Insecure           bool
	ConnectTimeout     time.Duration
	RequestTimeout     time.Duration
}

func Dial(cfg Config, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialContext(context.Background(), cfg, options...)
}

func DialContext(ctx context.Context, cfg Config, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Target) == "" {
		return nil, errors.New("gRPC target is required")
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5 * time.Second
	}
	transport := credentials.TransportCredentials(insecure.NewCredentials())
	if !cfg.Insecure {
		transport = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12, ServerName: cfg.ServerName})
	}
	options = append([]grpc.DialOption{
		grpc.WithTransportCredentials(transport),
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.DefaultConfig, MinConnectTimeout: cfg.ConnectTimeout}),
		grpc.WithChainUnaryInterceptor(PropagationUnaryClientInterceptor(cfg.RequestTimeout)),
	}, options...)
	return grpc.NewClient(cfg.Target, options...)
}
