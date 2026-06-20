import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  createSearchSynonym,
  deleteSearchSynonym,
  listSearchSynonyms,
  updateSearchSynonym
} from "../api/search-synonyms-api";
import type { SearchSynonymMutationInput } from "../types";

export const SEARCH_SYNONYMS_QUERY_KEY = ["admin", "search", "synonyms"] as const;

export function useSearchSynonyms() {
  return useQuery({
    queryKey: SEARCH_SYNONYMS_QUERY_KEY,
    queryFn: listSearchSynonyms,
    staleTime: 60_000
  });
}

export function useCreateSearchSynonym() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createSearchSynonym,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: SEARCH_SYNONYMS_QUERY_KEY });
    }
  });
}

export function useUpdateSearchSynonym() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ synonymId, input }: { synonymId: string; input: SearchSynonymMutationInput }) =>
      updateSearchSynonym(synonymId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: SEARCH_SYNONYMS_QUERY_KEY });
    }
  });
}

export function useDeleteSearchSynonym() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ synonymId, reason }: { synonymId: string; reason: string }) =>
      deleteSearchSynonym(synonymId, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: SEARCH_SYNONYMS_QUERY_KEY });
    }
  });
}
