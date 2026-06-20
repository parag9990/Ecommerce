import { CheckCircle2, RotateCcw, ShieldAlert, XCircle } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { ActionReasonDialog, type ReasonOption } from "../../../components/ui/action-reason-dialog";
import { cn } from "../../../lib/classnames";
import type { AdminRole } from "../../../lib/admin-rbac";
import { canReviewSellerKyc, canSuspendSeller } from "../permissions";
import type { AdminSeller, SellerStatus } from "../types";
import { useUpdateSellerStatus } from "../hooks/use-update-seller-status";

const APPROVE_REASON_CODES: readonly ReasonOption[] = [
  { value: "kyc_verified", label: "KYC verified" },
  { value: "business_verified", label: "Business verified" }
];

const REJECT_REASON_CODES: readonly ReasonOption[] = [
  { value: "kyc_invalid", label: "KYC invalid" },
  { value: "business_mismatch", label: "Business mismatch" }
];

const SUSPEND_REASON_CODES: readonly ReasonOption[] = [
  { value: "policy_violation", label: "Policy violation" },
  { value: "counterfeit_risk", label: "Counterfeit risk" },
  { value: "risk_review", label: "Risk review" }
];

const UNSUSPEND_REASON_CODES: readonly ReasonOption[] = [
  { value: "appeal_approved", label: "Appeal approved" },
  { value: "risk_cleared", label: "Risk cleared" }
];

type SellerAction = {
  key: "approve" | "reject" | "suspend" | "unsuspend";
  label: string;
  targetStatus: Extract<SellerStatus, "active" | "suspended" | "rejected">;
  title: string;
  description: string;
  confirmLabel: string;
  reasonLabel: string;
  tone: "neutral" | "danger";
  icon: LucideIcon;
  reasonOptions: readonly ReasonOption[];
};

const actionButtonClassName: Record<SellerAction["key"], string> = {
  approve: "border-emerald-200 bg-emerald-50 text-emerald-800 hover:bg-emerald-100",
  reject: "border-red-200 bg-red-50 text-red-800 hover:bg-red-100",
  suspend: "border-red-200 bg-red-50 text-red-800 hover:bg-red-100",
  unsuspend: "border-slate-300 bg-white text-slate-900 hover:bg-slate-50"
};

function buildSellerActions(seller: AdminSeller, roles: readonly AdminRole[]): SellerAction[] {
  const actions: SellerAction[] = [];

  if (seller.status === "pending_review" && canReviewSellerKyc(roles)) {
    actions.push(
      {
        key: "approve",
        label: "Approve seller",
        targetStatus: "active",
        title: "Approve seller",
        description: "Confirm KYC and business verification before activating this seller.",
        confirmLabel: "Approve",
        reasonLabel: "Approval reason",
        tone: "neutral",
        icon: CheckCircle2,
        reasonOptions: APPROVE_REASON_CODES
      },
      {
        key: "reject",
        label: "Reject seller",
        targetStatus: "rejected",
        title: "Reject seller",
        description: "Add a rejection reason that support teams can audit later.",
        confirmLabel: "Reject",
        reasonLabel: "Rejection reason",
        tone: "danger",
        icon: XCircle,
        reasonOptions: REJECT_REASON_CODES
      }
    );
  }

  if (seller.status === "active" && canSuspendSeller(roles)) {
    actions.push({
      key: "suspend",
      label: "Suspend seller",
      targetStatus: "suspended",
      title: "Suspend seller",
      description: "Choose a reason code and explain why seller access should be restricted.",
      confirmLabel: "Suspend",
      reasonLabel: "Suspension reason",
      tone: "danger",
      icon: ShieldAlert,
      reasonOptions: SUSPEND_REASON_CODES
    });
  }

  if (seller.status === "suspended" && canSuspendSeller(roles)) {
    actions.push({
      key: "unsuspend",
      label: "Unsuspend seller",
      targetStatus: "active",
      title: "Unsuspend seller",
      description: "Choose a reason code and explain why seller access can be restored.",
      confirmLabel: "Unsuspend",
      reasonLabel: "Unsuspension reason",
      tone: "neutral",
      icon: RotateCcw,
      reasonOptions: UNSUSPEND_REASON_CODES
    });
  }

  return actions;
}

function buildAuditReason(reasonCode: string | undefined, reason: string): string {
  return reasonCode ? `${reasonCode}: ${reason}` : reason;
}

export function SellerActionBar({ seller, roles }: { seller: AdminSeller; roles: readonly AdminRole[] }) {
  const [activeActionKey, setActiveActionKey] = useState<SellerAction["key"] | null>(null);
  const mutation = useUpdateSellerStatus();
  const actions = useMemo(() => buildSellerActions(seller, roles), [roles, seller]);
  const activeAction = actions.find((action) => action.key === activeActionKey) ?? null;

  const closeDialog = () => {
    if (mutation.isPending) {
      return;
    }

    setActiveActionKey(null);
    mutation.reset();
  };

  const submitAction = async ({ reason, reasonCode }: { reason: string; reasonCode?: string }) => {
    if (!activeAction) {
      return;
    }

    await mutation.mutateAsync({
      sellerId: seller.seller_id,
      status: activeAction.targetStatus,
      reason: buildAuditReason(reasonCode, reason)
    });

    setActiveActionKey(null);
  };

  return (
    <aside className="rounded-lg border border-slate-200 bg-white p-4">
      <h2 className="text-base font-semibold text-slate-950">Seller Actions</h2>
      <p className="mt-1 text-sm text-slate-600">Lifecycle mutations require an auditable reason.</p>

      {actions.length === 0 ? (
        <p className="mt-4 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-600">
          No lifecycle actions are available for your role and this seller status.
        </p>
      ) : (
        <div className="mt-4 grid gap-2">
          {actions.map((action) => {
            const Icon = action.icon;

            return (
              <button
                key={action.key}
                type="button"
                onClick={() => {
                  mutation.reset();
                  setActiveActionKey(action.key);
                }}
                disabled={mutation.isPending}
                className={cn(
                  "inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border px-3 text-sm font-medium disabled:cursor-not-allowed disabled:opacity-60",
                  actionButtonClassName[action.key]
                )}
              >
                <Icon className="h-4 w-4" aria-hidden="true" />
                {action.label}
              </button>
            );
          })}
        </div>
      )}

      {activeAction ? (
        <ActionReasonDialog
          open
          title={activeAction.title}
          description={activeAction.description}
          confirmLabel={activeAction.confirmLabel}
          reasonLabel={activeAction.reasonLabel}
          reasonOptions={activeAction.reasonOptions}
          tone={activeAction.tone}
          isSubmitting={mutation.isPending}
          error={mutation.error}
          onCancel={closeDialog}
          onConfirm={submitAction}
        />
      ) : null}
    </aside>
  );
}
