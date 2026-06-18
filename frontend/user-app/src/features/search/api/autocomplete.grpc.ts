import type { CallOptions } from '@connectrpc/connect';
import type { AutocompleteSuggestion } from '@ecommerce/proto-client';

import { grpcClients, type GrpcClients } from '../../../lib/grpc-client';
import { normalizeGrpcError } from '../../../lib/grpc-errors';
import { ApiError } from '../../../lib/http';

export type AutocompleteSuggestionsResult = Readonly<{
  items: AutocompleteSuggestion[];
  suggestions: string[];
}>;

type AutocompleteOptions = Readonly<{
  client?: Pick<GrpcClients['search'], 'autocomplete'> | undefined;
  limit?: number | undefined;
  signal?: AbortSignal | undefined;
  timeoutMs?: number | undefined;
}>;

export const AUTOCOMPLETE_MIN_QUERY_LENGTH = 2;
export const DEFAULT_AUTOCOMPLETE_LIMIT = 8;
export const MAX_AUTOCOMPLETE_LIMIT = 20;

function normalizeLimit(limit: number | undefined): number {
  const resolvedLimit = limit ?? DEFAULT_AUTOCOMPLETE_LIMIT;

  if (
    !Number.isInteger(resolvedLimit) ||
    resolvedLimit < 1 ||
    resolvedLimit > MAX_AUTOCOMPLETE_LIMIT
  ) {
    throw new ApiError(
      `Autocomplete limit must be between 1 and ${MAX_AUTOCOMPLETE_LIMIT}.`,
      'VALIDATION_ERROR',
      400,
    );
  }

  return resolvedLimit;
}

function buildCallOptions(options: AutocompleteOptions): CallOptions {
  const callOptions: CallOptions = {};

  if (options.signal) {
    callOptions.signal = options.signal;
  }

  if (options.timeoutMs !== undefined) {
    callOptions.timeoutMs = options.timeoutMs;
  }

  return callOptions;
}

function toSuggestionText(item: AutocompleteSuggestion): string | undefined {
  const text = item.value.trim() || item.label.trim();

  return text.length > 0 ? text : undefined;
}

export async function getAutocompleteSuggestions(
  query: string,
  options: AutocompleteOptions = {},
): Promise<AutocompleteSuggestionsResult> {
  const normalizedQuery = query.trim();

  if (normalizedQuery.length < AUTOCOMPLETE_MIN_QUERY_LENGTH) {
    return {
      items: [],
      suggestions: [],
    };
  }

  const client = options.client ?? grpcClients.search;

  try {
    const response = await client.autocomplete(
      {
        limit: normalizeLimit(options.limit),
        query: normalizedQuery,
      },
      buildCallOptions(options),
    );

    return {
      items: response.suggestions,
      suggestions: response.suggestions
        .map(toSuggestionText)
        .filter((suggestion): suggestion is string => Boolean(suggestion)),
    };
  } catch (error) {
    throw normalizeGrpcError(error);
  }
}
