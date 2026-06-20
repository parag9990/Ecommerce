import { Plus } from "lucide-react";
import type { FormEvent } from "react";

import type { SearchSynonymInput } from "../types";
import { normalizeSearchSynonymInput, validateSearchSynonymInput } from "../validators";

export function SearchSynonymForm({
  draft,
  existingRoots,
  canWrite,
  isSubmitting,
  onDraftChange,
  onSubmit
}: {
  draft: SearchSynonymInput;
  existingRoots: readonly string[];
  canWrite: boolean;
  isSubmitting?: boolean;
  onDraftChange: (draft: SearchSynonymInput) => void;
  onSubmit: (input: SearchSynonymInput) => void;
}) {
  const normalizedDraft = normalizeSearchSynonymInput(draft);
  const validation = validateSearchSynonymInput(draft);
  const duplicateRoot = Boolean(normalizedDraft.root && existingRoots.includes(normalizedDraft.root));
  const canSubmit = canWrite && validation.valid && !duplicateRoot && !isSubmitting;

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (canSubmit) {
      onSubmit(normalizedDraft);
    }
  }

  return (
    <form
      onSubmit={submit}
      className="grid gap-3 rounded-lg border border-slate-200 bg-slate-50 p-3 lg:grid-cols-[minmax(180px,0.8fr)_minmax(240px,1.5fr)_auto]"
    >
      <label>
        <span className="text-sm font-medium text-slate-700">Root term</span>
        <input
          value={draft.root}
          disabled={!canWrite || isSubmitting}
          onChange={(event) => onDraftChange({ ...draft, root: event.target.value })}
          placeholder="mobile"
          className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10 disabled:bg-slate-100 disabled:text-slate-500"
        />
      </label>

      <label>
        <span className="text-sm font-medium text-slate-700">Synonyms</span>
        <input
          value={draft.synonyms.join(", ")}
          disabled={!canWrite || isSubmitting}
          onChange={(event) =>
            onDraftChange({
              ...draft,
              synonyms: event.target.value.split(",")
            })
          }
          placeholder="phone, smartphone"
          className="mt-1 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10 disabled:bg-slate-100 disabled:text-slate-500"
        />
      </label>

      <div className="flex flex-col justify-end">
        <button
          type="submit"
          disabled={!canSubmit}
          className="inline-flex h-10 items-center justify-center gap-2 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
        >
          <Plus className="h-4 w-4" aria-hidden="true" />
          Add synonym
        </button>
      </div>

      {!validation.valid || duplicateRoot ? (
        <p className="lg:col-span-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {[...validation.errors, duplicateRoot ? "A synonym already exists for this root term." : ""]
            .filter(Boolean)
            .join(" ")}
        </p>
      ) : null}
    </form>
  );
}
