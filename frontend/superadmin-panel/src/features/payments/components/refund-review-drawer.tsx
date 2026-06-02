import { X } from "lucide-react";
import { useState } from "react";

import { ActionReasonDialog } from "../../../components/ui/action-reason-dialog";
import { AmountText } from "../../../components/ui/amount-text";
import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import { useRefundReview } from "../hooks/use-refund-review";
import type { Refund, RefundReviewDecision } from "../types";

type ReviewAction = {
  decision: RefundReviewDecision["decision"];
  label: string;
  title: string;
  description: string;
  tone: "neutral" | "danger";
};

const REVIEW_ACTIONS: readonly ReviewAction[] = [
  {
    decision: "rejected",
    label: "Reject refund",
    title: "Reject refund request",
    description: "Reject this refund and record a clear finance audit reason.",
    tone: "danger"
  },
  {
    decision: "approved",
    label: "Approve refund",
    title: "Approve refund request",
    description: "Approve this refund through the audited Superadmin review workflow.",
    tone: "neutral"
  }
] as const;

export function RefundReviewDrawer({
  refund,
  canReview,
  onClose
}: {
  refund: Refund | null;
  canReview: boolean;
  onClose: () => void;
}) {
  const [activeAction, setActiveAction] = useState<ReviewAction | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const mutation = useRefundReview();

  if (!refund) {
    return null;
  }

  const openAction = (action: ReviewAction) => {
    setSuccessMessage(null);
    mutation.reset();
    setActiveAction(action);
  };

  const closeDialog = () => {
    if (!mutation.isPending) {
      setActiveAction(null);
    }
  };

  const submitReview = async ({ reason }: { reason: string }) => {
    if (!activeAction) {
      return;
    }

    await mutation.mutateAsync({
      refundId: refund.refund_id,
      decision: activeAction.decision,
      reason
    });
    setSuccessMessage(`${activeAction.label} submitted.`);
    setActiveAction(null);
  };

  const reviewable = ["requested", "pending_review"].includes(refund.status);

  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        className="absolute inset-0 bg-slate-950/30"
        aria-label="Close refund review"
        onClick={onClose}
      />
      <aside className="absolute inset-y-0 right-0 flex w-full max-w-xl flex-col overflow-hidden border-l border-slate-200 bg-white shadow-xl">
        <header className="border-b border-slate-200 px-4 py-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h2 className="text-lg font-semibold text-slate-950">Refund Review</h2>
              <p className="mt-1 break-all text-sm text-slate-600">{refund.refund_id}</p>
            </div>
            <button
              type="button"
              onClick={onClose}
              title="Close refund review"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-100"
            >
              <X className="h-4 w-4" aria-hidden="true" />
              <span className="sr-only">Close refund review</span>
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-auto p-4">
          <section className="rounded-lg border border-slate-200 bg-slate-50 p-4">
            <dl className="grid gap-3 text-sm">
              <div className="flex items-center justify-between gap-3">
                <dt className="text-slate-600">Status</dt>
                <dd>
                  <StatusBadge status={refund.status} />
                </dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-600">Amount</dt>
                <dd className="font-medium text-slate-950">
                  <AmountText money={refund.amount} />
                </dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-600">Payment</dt>
                <dd className="break-all text-right font-medium text-slate-950">{refund.payment_id}</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-600">Order</dt>
                <dd className="break-all text-right text-slate-700">{refund.order_id ?? "Not linked"}</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-600">Requested by</dt>
                <dd className="capitalize text-slate-700">{refund.requested_by}</dd>
              </div>
              <div className="flex justify-between gap-3">
                <dt className="text-slate-600">Created</dt>
                <dd className="text-right text-slate-700">{formatDateTime(refund.created_at)}</dd>
              </div>
              <div className="grid gap-1">
                <dt className="text-slate-600">Request reason</dt>
                <dd className="rounded-lg border border-slate-200 bg-white px-3 py-2 text-slate-800">
                  {refund.reason}
                </dd>
              </div>
            </dl>
          </section>

          {!canReview ? (
            <p className="mt-4 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
              Your role can inspect refunds but cannot approve or reject them.
            </p>
          ) : !reviewable ? (
            <p className="mt-4 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
              This refund is no longer waiting for finance review.
            </p>
          ) : (
            <div className="mt-4 grid gap-2 sm:grid-cols-2">
              {REVIEW_ACTIONS.map((action) => (
                <button
                  key={action.decision}
                  type="button"
                  disabled={mutation.isPending}
                  onClick={() => openAction(action)}
                  className={
                    action.tone === "danger"
                      ? "h-10 rounded-lg border border-red-200 bg-red-50 px-3 text-sm font-medium text-red-800 hover:bg-red-100 disabled:opacity-60"
                      : "h-10 rounded-lg bg-slate-950 px-3 text-sm font-medium text-white hover:bg-slate-800 disabled:bg-slate-400"
                  }
                >
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
        </div>
      </aside>

      {activeAction ? (
        <ActionReasonDialog
          open
          title={activeAction.title}
          description={activeAction.description}
          confirmLabel={activeAction.label}
          reasonLabel="Finance review reason"
          tone={activeAction.tone}
          isSubmitting={mutation.isPending}
          error={mutation.error}
          onCancel={closeDialog}
          onConfirm={(input) => void submitReview(input)}
        />
      ) : null}
    </div>
  );
}
