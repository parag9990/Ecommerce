package grpctransport

import (
	"context"
	"errors"
	"log/slog"

	recommendationv1 "github.com/example/ecommerce-platform/backend/proto-gen/go/ecommerce/recommendation/v1"
	"github.com/example/ecommerce-platform/backend/services/recommendation-service/internal/usecase"
)

type Handler struct {
	recommendationv1.UnimplementedRecommendationServiceServer

	reader usecase.RecommendationReader
	logger *slog.Logger
}

func NewHandler(reader usecase.RecommendationReader, logger *slog.Logger) (*Handler, error) {
	if reader == nil {
		return nil, errors.New("recommendation reader is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{reader: reader, logger: logger}, nil
}

func (h *Handler) GetRecommendations(
	ctx context.Context,
	req *recommendationv1.GetRecommendationsRequest,
) (*recommendationv1.GetRecommendationsResponse, error) {
	input, err := inputFromProto(req, requestIDFromContext(ctx), sessionIDFromContext(ctx))
	if err != nil {
		return nil, grpcError(err)
	}

	output, err := h.reader.GetRecommendations(ctx, input)
	if err != nil {
		h.logger.WarnContext(ctx, "recommendation.grpc.get_recommendations_failed",
			slog.String("request_id", input.RequestID),
			slog.String("error", err.Error()),
		)
		return nil, grpcError(err)
	}
	return responseFromDomain(output), nil
}
