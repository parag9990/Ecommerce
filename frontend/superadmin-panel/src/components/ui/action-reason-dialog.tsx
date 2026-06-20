import { useEffect, useMemo, useState, type ReactNode } from "react";

import { ConfirmDialog } from "./confirm-dialog";

const DEFAULT_MIN_REASON_LENGTH = 10;

export type ReasonOption = {
  value: string;
  label: string;
};

export function ActionReasonDialog({
  open,
  title,
  description,
  confirmLabel,
  reasonLabel = "Audit reason",
  minLength = DEFAULT_MIN_REASON_LENGTH,
  tone = "neutral",
  reasonOptions,
  isSubmitting = false,
  error,
  children,
  onCancel,
  onConfirm
}: {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  reasonLabel?: string;
  minLength?: number;
  tone?: "neutral" | "danger";
  reasonOptions?: readonly ReasonOption[];
  isSubmitting?: boolean;
  error?: unknown;
  children?: ReactNode;
  onCancel: () => void;
  onConfirm: (input: { reason: string; reasonCode?: string }) => void | Promise<void>;
}) {
  const [reason, setReason] = useState("");
  const firstReasonCode = reasonOptions?.[0]?.value ?? "";
  const [reasonCode, setReasonCode] = useState(firstReasonCode);
  const reasonError = error instanceof Error ? error.message : error ? "Action failed." : null;
  const trimmedReason = reason.trim();
  const needsReasonCode = Boolean(reasonOptions?.length);
  const canSubmit =
    trimmedReason.length >= minLength && (!needsReasonCode || reasonCode.length > 0) && !isSubmitting;

  useEffect(() => {
    if (open) {
      setReason("");
      setReasonCode(firstReasonCode);
    }
  }, [firstReasonCode, open]);

  const selectedReasonLabel = useMemo(
    () => reasonOptions?.find((option) => option.value === reasonCode)?.label,
    [reasonCode, reasonOptions]
  );

  return (
    <ConfirmDialog
      open={open}
      title={title}
      description={description}
      tone={tone}
      onClose={onCancel}
      footer={
        <>
          <button
            type="button"
            onClick={onCancel}
            disabled={isSubmitting}
            className="h-10 rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-700 disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={!canSubmit}
            onClick={() => onConfirm({ reason: trimmedReason, reasonCode })}
            className="h-10 rounded-md bg-slate-950 px-3 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-400"
          >
            {isSubmitting ? "Submitting..." : confirmLabel}
          </button>
        </>
      }
    >
      {children ? <div className="mb-4">{children}</div> : null}

      {reasonOptions?.length ? (
        <label className="mb-3 block">
          <span className="text-sm font-medium text-slate-700">Reason code</span>
          <select
            value={reasonCode}
            onChange={(event) => setReasonCode(event.target.value)}
            className="mt-1 h-10 w-full rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-950 outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
          >
            {reasonOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      ) : null}

      <label className="block">
        <span className="text-sm font-medium text-slate-700">{reasonLabel}</span>
        <textarea
          value={reason}
          onChange={(event) => setReason(event.target.value)}
          rows={4}
          className="mt-1 w-full resize-none rounded-lg border border-slate-300 p-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
          placeholder="Add a clear reason for the audit trail"
        />
      </label>
      <div className="mt-2 text-xs text-slate-500">
        Minimum {minLength} characters.
        {selectedReasonLabel ? ` Selected code: ${selectedReasonLabel}.` : ""}
      </div>
      {reasonError ? (
        <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {reasonError}
        </p>
      ) : null}
    </ConfirmDialog>
  );
}
