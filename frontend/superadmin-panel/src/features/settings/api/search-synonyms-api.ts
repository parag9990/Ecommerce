import { apiFetch } from "../../../lib/http";
import type {
  SearchSynonym,
  SearchSynonymInput,
  SearchSynonymListResponse,
  SearchSynonymMutationInput
} from "../types";
import { normalizeSearchSynonymInput } from "../validators";

const SEARCH_SYNONYMS_PATH = "/api/v1/admin/search/synonyms";

type RawSearchSynonym = Partial<SearchSynonym>;
type RawSearchSynonymListResponse = {
  synonyms?: RawSearchSynonym[] | null;
};

function normalizeSynonym(item: RawSearchSynonym, index = 0): SearchSynonym {
  const normalizedInput = normalizeSearchSynonymInput({
    root: item.root ?? "",
    synonyms: item.synonyms ?? []
  });

  return {
    synonym_id: item.synonym_id?.trim() || `synonym_${index + 1}`,
    root: normalizedInput.root,
    synonyms: normalizedInput.synonyms,
    updated_at: item.updated_at ?? null
  };
}

export async function listSearchSynonyms(): Promise<SearchSynonymListResponse> {
  const response = await apiFetch<RawSearchSynonymListResponse>(SEARCH_SYNONYMS_PATH);

  return {
    synonyms: (response.synonyms ?? []).map(normalizeSynonym)
  };
}

export async function createSearchSynonym(
  input: SearchSynonymMutationInput
): Promise<SearchSynonym> {
  const payload = {
    ...normalizeSearchSynonymInput(input),
    reason: input.reason.trim()
  };
  const response = await apiFetch<RawSearchSynonym>(SEARCH_SYNONYMS_PATH, {
    method: "POST",
    body: JSON.stringify(payload)
  });

  return normalizeSynonym(response);
}

export async function updateSearchSynonym(
  synonymId: string,
  input: SearchSynonymMutationInput
): Promise<SearchSynonym> {
  const payload = {
    ...normalizeSearchSynonymInput(input),
    reason: input.reason.trim()
  };
  const response = await apiFetch<RawSearchSynonym>(
    `${SEARCH_SYNONYMS_PATH}/${encodeURIComponent(synonymId)}`,
    {
      method: "PATCH",
      body: JSON.stringify(payload)
    }
  );

  return normalizeSynonym({ ...response, synonym_id: response.synonym_id ?? synonymId });
}

export async function deleteSearchSynonym(
  synonymId: string,
  reason: string
): Promise<{ success: boolean }> {
  return apiFetch<{ success: boolean }>(
    `${SEARCH_SYNONYMS_PATH}/${encodeURIComponent(synonymId)}`,
    {
      method: "DELETE",
      body: JSON.stringify({ reason: reason.trim() })
    }
  );
}
