import type {
  AutocompleteResponse,
} from '@ecommerce/proto-client/gen/ecommerce/search/v1/search_pb';
import { describe, expect, it, vi } from 'vitest';

import type { GrpcClients } from '../../../lib/grpc-client';
import { getAutocompleteSuggestions } from './autocomplete.grpc';

describe('getAutocompleteSuggestions', () => {
  it('does not call gRPC for short queries', async () => {
    const autocomplete = vi.fn<GrpcClients['search']['autocomplete']>();
    const client: Pick<GrpcClients['search'], 'autocomplete'> = { autocomplete };

    const result = await getAutocompleteSuggestions('s', { client });

    expect(result.suggestions).toEqual([]);
    expect(autocomplete).not.toHaveBeenCalled();
  });

  it('normalizes typed autocomplete suggestions for the UI', async () => {
    const autocomplete = vi
      .fn<GrpcClients['search']['autocomplete']>()
      .mockResolvedValue({
        $typeName: 'ecommerce.search.v1.AutocompleteResponse',
        suggestions: [
          {
            $typeName: 'ecommerce.search.v1.AutocompleteSuggestion',
            label: 'Running Shoes',
            type: 'product',
            value: 'running shoes',
          },
        ],
      } satisfies AutocompleteResponse);
    const client: Pick<GrpcClients['search'], 'autocomplete'> = { autocomplete };

    const result = await getAutocompleteSuggestions(' running ', { client });

    expect(result.suggestions).toEqual(['running shoes']);
    expect(autocomplete).toHaveBeenCalledWith(
      {
        limit: 8,
        query: 'running',
      },
      {},
    );
  });
});
