import { useState } from "react";
import { Edit2, RotateCcw, Save, Trash2, X } from "lucide-react";

import { ActionReasonDialog } from "../../../components/ui/action-reason-dialog";
import { DataState } from "../../../components/ui/data-state";
import { formatDateTime } from "../../../lib/format";
import {
  useCreateSearchSynonym,
  useDeleteSearchSynonym,
  useSearchSynonyms,
  useUpdateSearchSynonym
} from "../hooks/use-search-synonyms";
import type { SearchSynonym, SearchSynonymInput } from "../types";
import { normalizeSearchSynonymInput, validateSearchSynonymInput } from "../validators";
import { SearchSynonymForm } from "./search-synonym-form";
import { SettingChangeSummary } from "./setting-change-summary";

type PendingSynonymAction =
  | { type: "create"; input: SearchSynonymInput }
  | { type: "update"; synonym: SearchSynonym; input: SearchSynonymInput }
  | { type: "delete"; synonym: SearchSynonym }
  | null;

const EMPTY_DRAFT: SearchSynonymInput = {
  root: "",
  synonyms: []
};

function serializeSynonym(input: SearchSynonymInput): string {
  const normalized = normalizeSearchSynonymInput(input);

  return JSON.stringify({
    root: normalized.root,
    synonyms: [...normalized.synonyms].sort()
  });
}

function toDraft(synonym: SearchSynonym): SearchSynonymInput {
  return {
    root: synonym.root,
    synonyms: synonym.synonyms
  };
}

function mutationError(
  pendingAction: PendingSynonymAction,
  createError: unknown,
  updateError: unknown,
  deleteError: unknown
): unknown {
  if (pendingAction?.type === "create") {
    return createError;
  }

  if (pendingAction?.type === "update") {
    return updateError;
  }

  if (pendingAction?.type === "delete") {
    return deleteError;
  }

  return null;
}

export function SearchSynonymTable({ canWrite }: { canWrite: boolean }) {
  const synonymsQuery = useSearchSynonyms();
  const createMutation = useCreateSearchSynonym();
  const updateMutation = useUpdateSearchSynonym();
  const deleteMutation = useDeleteSearchSynonym();
  const [createDraft, setCreateDraft] = useState<SearchSynonymInput>(EMPTY_DRAFT);
  const [editSynonymId, setEditSynonymId] = useState<string | null>(null);
  const [editDraft, setEditDraft] = useState<SearchSynonymInput>(EMPTY_DRAFT);
  const [pendingAction, setPendingAction] = useState<PendingSynonymAction>(null);
  const isMutating =
    createMutation.isPending || updateMutation.isPending || deleteMutation.isPending;
  const synonyms = synonymsQuery.data?.synonyms ?? [];
  const existingRoots = synonyms.map((synonym) => synonym.root);
  const activeError = mutationError(
    pendingAction,
    createMutation.error,
    updateMutation.error,
    deleteMutation.error
  );

  function resetMutations() {
    createMutation.reset();
    updateMutation.reset();
    deleteMutation.reset();
  }

  function startEdit(synonym: SearchSynonym) {
    resetMutations();
    setEditSynonymId(synonym.synonym_id);
    setEditDraft(toDraft(synonym));
  }

  function cancelEdit() {
    setEditSynonymId(null);
    setEditDraft(EMPTY_DRAFT);
  }

  async function submitPendingAction({ reason }: { reason: string }) {
    if (!pendingAction) {
      return;
    }

    try {
      if (pendingAction.type === "create") {
        await createMutation.mutateAsync({ ...pendingAction.input, reason });
        setCreateDraft(EMPTY_DRAFT);
      } else if (pendingAction.type === "update") {
        await updateMutation.mutateAsync({
          synonymId: pendingAction.synonym.synonym_id,
          input: { ...pendingAction.input, reason }
        });
        cancelEdit();
      } else {
        await deleteMutation.mutateAsync({
          synonymId: pendingAction.synonym.synonym_id,
          reason
        });
      }

      setPendingAction(null);
    } catch {
      // ActionReasonDialog displays the active mutation error and preserves the draft.
    }
  }

  if (synonymsQuery.isLoading) {
    return <DataState title="Loading search synonyms" description="Fetching search config." />;
  }

  if (synonymsQuery.isError) {
    return (
      <DataState
        tone="danger"
        title="Search synonyms could not be loaded"
        description={
          synonymsQuery.error instanceof Error ? synonymsQuery.error.message : "Search API request failed."
        }
        action={
          <button
            type="button"
            onClick={() => void synonymsQuery.refetch()}
            className="h-9 rounded-lg border border-red-200 bg-white px-3 text-sm font-medium text-red-700"
          >
            Retry
          </button>
        }
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="rounded-lg border border-slate-200 bg-white p-4">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Search Synonyms</h2>
          <p className="mt-1 text-sm text-slate-600">
            Manage equivalent search terms that affect product discovery.
          </p>
        </div>

        <div className="mt-4">
          <SearchSynonymForm
            draft={createDraft}
            existingRoots={existingRoots}
            canWrite={canWrite}
            isSubmitting={isMutating}
            onDraftChange={setCreateDraft}
            onSubmit={(input) => {
              resetMutations();
              setPendingAction({ type: "create", input });
            }}
          />
        </div>

        {!canWrite ? (
          <div className="mt-4">
            <DataState
              title="Read-only search synonyms"
              description="Superadmins and catalog admins can manage synonyms."
            />
          </div>
        ) : null}
      </div>

      <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
        <div className="border-b border-slate-200 px-4 py-3">
          <h3 className="text-sm font-semibold text-slate-950">Configured synonyms</h3>
          <p className="mt-1 text-xs text-slate-500">{synonyms.length} root terms configured.</p>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-slate-50 text-xs uppercase text-slate-500">
              <tr>
                <th className="w-56 px-4 py-3 font-medium">Root</th>
                <th className="px-4 py-3 font-medium">Synonyms</th>
                <th className="w-44 px-4 py-3 font-medium">Updated</th>
                <th className="w-32 px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {synonyms.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-4 py-6 text-sm text-slate-500">
                    No synonyms configured.
                  </td>
                </tr>
              ) : (
                synonyms.map((synonym) => {
                  const isEditing = editSynonymId === synonym.synonym_id;
                  const editValidation = validateSearchSynonymInput(editDraft);
                  const editDirty = serializeSynonym(editDraft) !== serializeSynonym(toDraft(synonym));
                  const rootExists = synonyms.some(
                    (item) =>
                      item.synonym_id !== synonym.synonym_id &&
                      item.root === normalizeSearchSynonymInput(editDraft).root
                  );
                  const canSaveEdit = canWrite && editValidation.valid && editDirty && !rootExists && !isMutating;

                  return (
                    <tr key={synonym.synonym_id}>
                      <td className="align-top px-4 py-3">
                        {isEditing ? (
                          <label>
                            <span className="sr-only">Edit root term</span>
                            <input
                              value={editDraft.root}
                              disabled={isMutating}
                              onChange={(event) =>
                                setEditDraft((current) => ({ ...current, root: event.target.value }))
                              }
                              className="h-9 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
                            />
                          </label>
                        ) : (
                          <span className="font-medium text-slate-950">{synonym.root}</span>
                        )}
                      </td>
                      <td className="align-top px-4 py-3">
                        {isEditing ? (
                          <div>
                            <label>
                              <span className="sr-only">Edit synonyms</span>
                              <input
                                value={editDraft.synonyms.join(", ")}
                                disabled={isMutating}
                                onChange={(event) =>
                                  setEditDraft((current) => ({
                                    ...current,
                                    synonyms: event.target.value.split(",")
                                  }))
                                }
                                className="h-9 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
                              />
                            </label>
                            {!editValidation.valid || rootExists ? (
                              <p className="mt-2 text-xs text-red-700">
                                {[...editValidation.errors, rootExists ? "Root term already exists." : ""]
                                  .filter(Boolean)
                                  .join(" ")}
                              </p>
                            ) : null}
                          </div>
                        ) : (
                          <div className="flex flex-wrap gap-2">
                            {synonym.synonyms.map((term) => (
                              <span
                                key={term}
                                className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-xs font-medium text-slate-700"
                              >
                                {term}
                              </span>
                            ))}
                          </div>
                        )}
                      </td>
                      <td className="align-top px-4 py-3 text-slate-600">
                        {formatDateTime(synonym.updated_at)}
                      </td>
                      <td className="align-top px-4 py-3">
                        <div className="flex justify-end gap-2">
                          {isEditing ? (
                            <>
                              <button
                                type="button"
                                onClick={cancelEdit}
                                disabled={isMutating}
                                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                                aria-label="Cancel synonym edit"
                              >
                                <X className="h-4 w-4" aria-hidden="true" />
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  resetMutations();
                                  setPendingAction({
                                    type: "update",
                                    synonym,
                                    input: normalizeSearchSynonymInput(editDraft)
                                  });
                                }}
                                disabled={!canSaveEdit}
                                className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-slate-950 text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                                aria-label={`Save synonym ${synonym.root}`}
                              >
                                <Save className="h-4 w-4" aria-hidden="true" />
                              </button>
                            </>
                          ) : (
                            <>
                              <button
                                type="button"
                                onClick={() => startEdit(synonym)}
                                disabled={!canWrite || isMutating}
                                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                                aria-label={`Edit synonym ${synonym.root}`}
                              >
                                <Edit2 className="h-4 w-4" aria-hidden="true" />
                              </button>
                              <button
                                type="button"
                                onClick={() => {
                                  resetMutations();
                                  setPendingAction({ type: "delete", synonym });
                                }}
                                disabled={!canWrite || isMutating}
                                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-red-200 text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50"
                                aria-label={`Delete synonym ${synonym.root}`}
                              >
                                <Trash2 className="h-4 w-4" aria-hidden="true" />
                              </button>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>

      <ActionReasonDialog
        open={pendingAction !== null}
        title={
          pendingAction?.type === "delete"
            ? "Delete search synonym"
            : pendingAction?.type === "update"
              ? "Update search synonym"
              : "Create search synonym"
        }
        description="Confirm the search synonym change and provide an audit reason."
        confirmLabel="Confirm update"
        tone={pendingAction?.type === "delete" ? "danger" : "neutral"}
        isSubmitting={isMutating}
        error={activeError}
        onCancel={() => {
          if (!isMutating) {
            setPendingAction(null);
            resetMutations();
          }
        }}
        onConfirm={submitPendingAction}
      >
        <SettingChangeSummary
          title="Change summary"
          tone={pendingAction?.type === "delete" ? "warning" : "neutral"}
          items={[
            {
              label: "Action",
              value: pendingAction?.type ?? "unknown"
            },
            {
              label: "Root",
              value:
                pendingAction?.type === "delete"
                  ? pendingAction.synonym.root
                  : pendingAction?.input.root ?? "Unknown"
            },
            {
              label: "Synonyms",
              value:
                pendingAction?.type === "delete"
                  ? pendingAction.synonym.synonyms.join(", ")
                  : pendingAction?.input.synonyms.join(", ") || "None"
            }
          ]}
        />
      </ActionReasonDialog>
    </section>
  );
}
