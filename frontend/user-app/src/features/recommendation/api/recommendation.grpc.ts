import type { CallOptions } from '@connectrpc/connect';
import type { GetRecommendationsResponse } from '@ecommerce/proto-client/gen/ecommerce/recommendation/v1/recommendation_pb';

import { grpcClients, type GrpcClients } from '../../../lib/grpc-client';
import { normalizeGrpcError } from '../../../lib/grpc-errors';
import { ApiError } from '../../../lib/http';

export type RecommendationPageType = 'cart' | 'home' | 'product_detail';

export type RecommendationContextInput = Readonly<{
  categoryId?: string | undefined;
  pageType: RecommendationPageType;
  productId?: string | undefined;
  userId?: string | undefined;
}>;

type RecommendationOptions = Readonly<{
  client?: Pick<GrpcClients['recommendation'], 'getRecommendations'> | undefined;
  limit?: number | undefined;
  signal?: AbortSignal | undefined;
  timeoutMs?: number | undefined;
}>;

export const DEFAULT_RECOMMENDATION_LIMIT = 12;
export const MAX_RECOMMENDATION_LIMIT = 48;

function normalizeLimit(limit: number | undefined): number {
  const resolvedLimit = limit ?? DEFAULT_RECOMMENDATION_LIMIT;

  if (
    !Number.isInteger(resolvedLimit) ||
    resolvedLimit < 1 ||
    resolvedLimit > MAX_RECOMMENDATION_LIMIT
  ) {
    throw new ApiError(
      `Recommendation limit must be between 1 and ${MAX_RECOMMENDATION_LIMIT}.`,
      'VALIDATION_ERROR',
      400,
    );
  }

  return resolvedLimit;
}

function buildCallOptions(options: RecommendationOptions): CallOptions {
  const callOptions: CallOptions = {};

  if (options.signal) {
    callOptions.signal = options.signal;
  }

  if (options.timeoutMs !== undefined) {
    callOptions.timeoutMs = options.timeoutMs;
  }

  return callOptions;
}

export async function getRecommendations(
  context: RecommendationContextInput,
  options: RecommendationOptions = {},
): Promise<GetRecommendationsResponse> {
  const client = options.client ?? grpcClients.recommendation;

  try {
    return await client.getRecommendations(
      {
        context: {
          categoryId: context.categoryId ?? '',
          pageType: context.pageType,
          productId: context.productId ?? '',
          userId: context.userId ?? '',
        },
        limit: normalizeLimit(options.limit),
      },
      buildCallOptions(options),
    );
  } catch (error) {
    throw normalizeGrpcError(error);
  }
}
