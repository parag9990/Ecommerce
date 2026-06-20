import { ShieldCheck, ShieldAlert } from "lucide-react";
import { useState } from "react";

import { ConfirmDialog } from "../../../components/ui/confirm-dialog";
import { StatusBadge } from "../../../components/ui/status-badge";
import type { AdminRole } from "../../../lib/admin-rbac";
import { cn } from "../../../lib/classnames";
import { canReviewOrders } from "../permissions";
import { useSubmitOrderReview } from "../hooks/use-submit-order-review";
import type { ManualReviewDecision, ReviewStatus } from "../types";

const MIN_REASON_LENGTH = 10;

type ReviewAction = {
  decision: ManualReviewDecision["decision"];
  label: string;
  title: string;
  description: string;
  tone: "neutral" | "danger";
};

const REVIEW_ACTIONS: readonly ReviewAction[] = [
  {
    decision: "mark_reviewing",
    label: "Mark reviewing",
    title: "Mark order for manual review",
    description: "Use this when fulfillment, dispute, or fraud signals need admin investigation.",
    tone: "neutral"
  },
  {
    decision: "resolve",
    label: "Resolve review",
    title: "Resolve manual review",
    description: "Record why this order no longer requires admin review.",
    tone: "neutral"
  },
  {
    decision: "escalate",
    label: "Escalate",
    title: "Escalate order review",
    description: "Escalate this order to a higher support or risk workflow with an auditable reason.",
    tone: "danger"
  }
] as const;

function actionButtonClassName(action: ReviewAction): string {
  if (action.tone === "danger") {
    return "border-red-200 bg-red-50 text-red-800 hover:bg-red-100";
  }

  if (action.decision === "resolve") {
    return "border-emerald-200 bg-emerald-50 text-emerald-800 hover:bg-emerald-100";
  }

  return "border-slate-300 bg-white text-slate-900 hover:bg-slate-50";
}

export function ManualReviewDrawer({
  orderId,
  reviewStatus,
  roles
}: {
  orderId: string;
  reviewStatus: ReviewStatus;
  roles: readonly AdminRole[];
}) {
  const [activeAction, setActiveAction] = useState<ReviewAction | null>(null);
  const [reason, setReason] = useState("");
  const [internalNote, setInternalNote] = useState("");
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const mutation = useSubmitOrderReview(orderId);
  const canReview = canReviewOrders(roles);
  const trimmedReason = reason.trim();
  const canSubmit = trimmedReason.length >= MIN_REASON_LENGTH && !mutation.isPending;

  const openAction = (action: ReviewAction) => {
    setReason("");
    setInternalNote("");
    setSuccessMessage(null);
    mutation.reset();
    setActiveAction(action);
  };

  const closeDialog = () => {
    if (mutation.isPending) {
      return;
    }

    setActiveAction(null);
  };

  const submitReview = async () => {
    if (!activeAction || !canSubmit) {
      return;
    }

    await mutation.mutateAsync({
      decision: activeAction.decision,
      reason: trimmedReason,
      internal_note: internalNote.trim() || undefined
    });

    setSuccessMessage(`${activeAction.label} submitted.`);
    setActiveAction(null);
  };

  return (
    <aside className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Manual Review</h2>
          <p className="mt-1 text-sm text-slate-600">Admin review decisions require a reason.</p>
        </div>
        <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-100 text-slate-600">
          <ShieldCheck className="h-5 w-5" aria-hidden="true" />
        </span>
      </div>

      <div className="mt-4 flex items-center justify-between gap-3 text-sm">
        <span className="text-slate-600">Current review state</span>
        <StatusBadge status={reviewStatus} />
      </div>

      {!canReview ? (
        <p className="mt-4 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
          Your role can inspect orders but cannot submit manual review decisions.
        </p>
      ) : (
        <div className="mt-4 grid gap-2">
          {REVIEW_ACTIONS.map((action) => (
            <button
              key={action.decision}
              type="button"
              onClick={() => openAction(action)}
              disabled={mutation.isPending}
              className={cn(
                "inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border px-3 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-60",
                actionButtonClassName(action)
              )}
            >
              <ShieldAlert className="h-4 w-4" aria-hidden="true" />
              {action.label}
            </button>
          ))}
        </div>
      )}

      {successMessage ? (
        <p className="mt-3 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
          {successMessage}
        </p>
      ) : null}

      {activeAction ? (
        <ConfirmDialog
          open
          title={activeAction.title}
          description={activeAction.description}
          tone={activeAction.tone}
          onClose={closeDialog}
          footer={
            <>
              <button
                type="button"
                onClick={closeDialog}
                disabled={mutation.isPending}
                className="h-10 rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-700 disabled:opacity-50"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={!canSubmit}
                onClick={() => void submitReview()}
                className="h-10 rounded-md bg-slate-950 px-3 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-400"
              >
                {mutation.isPending ? "Submitting..." : activeAction.label}
              </button>
            </>
          }
        >
          <label className="block">
            <span className="text-sm font-medium text-slate-700">Review reason</span>
            <textarea
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              rows={4}
              className="mt-1 w-full resize-none rounded-lg border border-slate-300 p-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
              placeholder="Add the operational reason for the audit trail"
            />
          </label>
          <label className="mt-3 block">
            <span className="text-sm font-medium text-slate-700">Internal admin note</span>
            <textarea
              value={internalNote}
              onChange={(event) => setInternalNote(event.target.value)}
              rows={3}
              className="mt-1 w-full resize-none rounded-lg border border-slate-300 p-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
              placeholder="Optional investigation note"
            />
          </label>
          <div className="mt-2 text-xs text-slate-500">
            Reason must be at least {MIN_REASON_LENGTH} characters. Internal note is optional.
          </div>
          {mutation.error ? (
            <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {mutation.error instanceof Error ? mutation.error.message : "Manual review submission failed."}
            </p>
          ) : null}
        </ConfirmDialog>
      ) : null}
    </aside>
  );
}
