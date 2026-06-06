package grpcweb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"ecommerce/api-gateway/internal/clients"
	"ecommerce/api-gateway/internal/config"
	"ecommerce/api-gateway/internal/observability"
	"ecommerce/api-gateway/internal/usecase"

	"google.golang.org/grpc"
)

type ServerOptions struct {
	Config        config.GRPCWebConfig
	ServiceName   string
	Observability observability.Config
	Policies      usecase.GRPCWebPolicyCatalog
	Connections   ConnectionProvider
	TokenVerifier TokenVerifier
	Logger        *slog.Logger
	Metrics       *observability.Metrics
	Listener      net.Listener
}

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(ctx context.Context, options ServerOptions) (*Server, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if options.Policies == nil {
		return nil, errors.New("grpc-web policy catalog is required")
	}
	if options.Connections == nil {
		return nil, errors.New("grpc-web downstream connections are required")
	}
	if options.Config.MaxReceiveMsgBytes <= 0 {
		return nil, errors.New("grpc-web max receive message bytes must be positive")
	}
	if options.Config.MaxSendMsgBytes <= 0 {
		return nil, errors.New("grpc-web max send message bytes must be positive")
	}
	if options.Listener == nil && options.Config.Address == "" {
		return nil, errors.New("grpc-web address is required")
	}
	requiresToken, err := options.Policies.RequiresToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("inspect grpc-web policies: %w", err)
	}
	if requiresToken && options.TokenVerifier == nil {
		return nil, errors.New("grpc-web token verifier is required by configured policies")
	}
	policies, err := options.Policies.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list grpc-web policies: %w", err)
	}
	for _, policy := range policies {
		if connection, ok := options.Connections.Connection(clients.Downstream(policy.Downstream)); !ok || connection == nil {
			return nil, fmt.Errorf("grpc-web policy %q references unavailable downstream %q", policy.FullMethod, policy.Downstream)
		}
	}

	listener := options.Listener
	if listener == nil {
		listener, err = net.Listen("tcp", options.Config.Address)
		if err != nil {
			return nil, fmt.Errorf("listen for grpc-web facade on %s: %w", options.Config.Address, err)
		}
	}
	proxy := newTransparentProxy(options.Connections)
	server := grpc.NewServer(
		grpc.ForceServerCodec(proxy.codec),
		grpc.MaxRecvMsgSize(options.Config.MaxReceiveMsgBytes),
		grpc.MaxSendMsgSize(options.Config.MaxSendMsgBytes),
		grpc.UnknownServiceHandler(proxy.Handler),
		grpc.ChainStreamInterceptor(securityInterceptor(
			options.ServiceName,
			options.Observability,
			options.Policies,
			options.TokenVerifier,
			options.Logger,
			options.Metrics,
		)),
	)
	return &Server{grpcServer: server, listener: listener}, nil
}

func (s *Server) Serve() error {
	if s == nil || s.grpcServer == nil || s.listener == nil {
		return errors.New("grpc-web server is not initialized")
	}
	err := s.grpcServer.Serve(s.listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.grpcServer == nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.grpcServer.Stop()
		return ctx.Err()
	}
}

func (s *Server) Address() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}
