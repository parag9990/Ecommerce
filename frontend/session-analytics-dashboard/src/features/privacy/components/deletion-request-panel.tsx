import { AlertTriangle, Search, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";

import {
  deletionTargetTypeOptions,
  type DeletionTargetType
} from "../../../api/session-api";
import { formatNumber } from "../../../lib/format";
import {
  useCreateDeletionRequest,
  usePreviewDeletion
} from "../hooks/use-deletion-requests";
import {
  validateDeletionReason,
  validateDeletionTargetValue
} from "../lib/privacy-validation";
import { PrivacyAuditNote } from "./privacy-audit-note";
import { SensitiveActionConfirmation } from "./sensitive-action-confirmation";

type DeletionRequestPanelProps = {
  canRequestDeletion: boolean;
};

export function DeletionRequestPanel({
  canRequestDeletion
}: DeletionRequestPanelProps) {
  const [targetType, setTargetType] = useState<DeletionTargetType>("user_id");
  const [targetValue, setTargetValue] = useState("");
  const [reason, setReason] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);
  const preview = usePreviewDeletion();
  const createRequest = useCreateDeletionRequest();
  const targetError = useMemo(
    () =>
      targetValue
        ? validateDeletionTargetValue(targetType, targetValue)
        : undefined,
    [targetType, targetValue]
  );
  const reasonError = reason ? validateDeletionReason(reason) : undefined;
  const canPreview =
    canRequestDeletion && Boolean(targetValue.trim()) && !targetError;
  const canSubmit =
    canRequestDeletion &&
    Boolean(preview.data) &&
    !targetError &&
    !validateDeletionReason(reason) &&
    !createRequest.isPending;

  useEffect(() => {
    preview.reset();
    setConfirmOpen(false);
  }, [targetType, targetValue]);

  return (
    <section className="rounded-lg border border-zinc-200 bg-white p-4 shadow-panel">
      <div>
        <p className="text-xs font-semibold uppercase text-red-700">
          High-risk action
        </p>
        <h2 className="mt-1 text-base font-semibold text-zinc-950">
          Deletion request
        </h2>
        <p className="mt-1 text-sm text-zinc-500">
          Preview impact before removing or anonymizing matching analytics data.
        </p>
      </div>

      {!canRequestDeletion ? (
        <div className="mt-4 rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-900">
          Deletion requests require an operations admin or superadmin role.
        </div>
      ) : null}

      <div className="mt-4 space-y-3">
        <label className="block text-sm font-medium text-zinc-700">
          Target type
          <select
            className="mt-1 h-10 w-full rounded-md border border-zinc-300 bg-white px-3 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100"
            disabled={!canRequestDeletion}
            onChange={(event) =>
              setTargetType(event.target.value as DeletionTargetType)
            }
            value={targetType}
          >
            {deletionTargetTypeOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="block text-sm font-medium text-zinc-700">
          Target value
          <input
            className={[
              "mt-1 h-10 w-full rounded-md border bg-white px-3 font-mono text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100",
              targetError ? "border-red-300" : "border-zinc-300"
            ].join(" ")}
            disabled={!canRequestDeletion}
            maxLength={160}
            onChange={(event) => setTargetValue(event.target.value)}
            placeholder="user_123, anon_123, or sess_123"
            value={targetValue}
          />
          {targetError ? (
            <span className="mt-1 block text-xs text-red-600">{targetError}</span>
          ) : null}
        </label>

        <button
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-md border border-zinc-300 bg-white px-3 text-sm font-medium text-zinc-800 transition-colors hover:bg-zinc-50 disabled:cursor-not-allowed disabled:text-zinc-400"
          disabled={!canPreview || preview.isPending}
          onClick={() =>
            preview.mutate({
              targetType,
              targetValue
            })
          }
          type="button"
        >
          <Search className="h-4 w-4" aria-hidden="true" />
          {preview.isPending ? "Previewing" : "Preview impact"}
        </button>

        {preview.isError ? (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            Deletion impact could not be previewed.
          </div>
        ) : null}

        {preview.data ? (
          <div className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
            <div className="flex gap-2">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              <div>
                <p className="font-medium">
                  Preview for {preview.data.targetValueMasked}
                </p>
                <div className="mt-2 grid gap-1 text-xs">
                  <span>{formatNumber(preview.data.matchedSessions)} sessions</span>
                  <span>{formatNumber(preview.data.matchedEvents)} events</span>
                  <span>
                    {formatNumber(preview.data.matchedJourneySummaries)} journey
                    summaries
                  </span>
                  <span>
                    {formatNumber(preview.data.matchedActiveSessions)} active
                    sessions
                  </span>
                </div>
              </div>
            </div>
          </div>
        ) : null}

        <PrivacyAuditNote
          disabled={!canRequestDeletion}
          error={reasonError}
          id="deletion-reason"
          onChange={setReason}
          value={reason}
        />

        {createRequest.isError ? (
          <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
            Deletion request could not be submitted.
          </div>
        ) : null}

        {createRequest.data ? (
          <div className="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
            Deletion request {createRequest.data.requestId} is{" "}
            {createRequest.data.status}.
          </div>
        ) : null}

        <button
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-md bg-red-600 px-3 text-sm font-medium text-white transition-colors hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-zinc-400"
          disabled={!canSubmit}
          onClick={() => setConfirmOpen(true)}
          type="button"
        >
          <Trash2 className="h-4 w-4" aria-hidden="true" />
          Submit deletion request
        </button>
      </div>

      <SensitiveActionConfirmation
        confirmLabel="Submit request"
        description="This will remove raw matching analytics data and anonymize linked session records. The audit reason will be stored with the request."
        isBusy={createRequest.isPending}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={() =>
          createRequest.mutate(
            {
              confirmed: true,
              reason,
              targetType,
              targetValue
            },
            {
              onSuccess: () => {
                setConfirmOpen(false);
                setReason("");
              }
            }
          )
        }
        open={confirmOpen}
        title="Confirm analytics deletion"
      />
    </section>
  );
}
