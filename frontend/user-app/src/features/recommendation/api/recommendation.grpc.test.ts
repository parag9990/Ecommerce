import type {
  GetRecommendationsResponse,
} from '@ecommerce/proto-client/gen/ecommerce/recommendation/v1/recommendation_pb';
import { describe, expect, it, vi } from 'vitest';

import type { GrpcClients } from '../../../lib/grpc-client';
import { getRecommendations } from './recommendation.grpc';

describe('getRecommendations', () => {
  it('sends normalized recommendation context through the typed client', async () => {
    const getRecommendationsRpc = vi
      .fn<GrpcClients['recommendation']['getRecommendations']>()
      .mockResolvedValue({
        $typeName: 'ecommerce.recommendation.v1.GetRecommendationsResponse',
        items: [
          {
            $typeName: 'ecommerce.recommendation.v1.RecommendationItem',
            imageUrl: '/products/related.jpg',
            priceDisplay: '$49.00',
            productId: 'prod_related',
            reason: 'Frequently viewed together',
            title: 'Related product',
          },
        ],
        requestId: 'req_test',
      } satisfies GetRecommendationsResponse);
    const client: Pick<GrpcClients['recommendation'], 'getRecommendations'> = {
      getRecommendations: getRecommendationsRpc,
    };

    const response = await getRecommendations(
      {
        categoryId: 'cat_shoes',
        pageType: 'product_detail',
        productId: 'prod_123',
      },
      { client },
    );

    expect(response.items[0]?.productId).toBe('prod_related');
    expect(getRecommendationsRpc).toHaveBeenCalledWith(
      {
        context: {
          categoryId: 'cat_shoes',
          pageType: 'product_detail',
          productId: 'prod_123',
          userId: '',
        },
        limit: 12,
      },
      {},
    );
  });

  it('rejects unsafe recommendation limits before making an RPC', async () => {
    const getRecommendationsRpc =
      vi.fn<GrpcClients['recommendation']['getRecommendations']>();
    const client: Pick<GrpcClients['recommendation'], 'getRecommendations'> = {
      getRecommendations: getRecommendationsRpc,
    };

    await expect(
      getRecommendations({ pageType: 'home' }, { client, limit: 0 }),
    ).rejects.toMatchObject({
      code: 'VALIDATION_ERROR',
      status: 400,
    });
    expect(getRecommendationsRpc).not.toHaveBeenCalled();
  });
});
