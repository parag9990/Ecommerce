import { Send, UserX } from "lucide-react";
import { useState } from "react";

import { getSafeErrorMessage } from "../../../lib/api-error";
import type { SellerStaffMember } from "../types";
import { useSellerPermissions } from "../hooks/use-seller-permissions";
import {
  useDisableSellerStaff,
  useResendSellerStaffInvite,
} from "../hooks/use-seller-team";

type StaffActionsMenuProps = {
  member: SellerStaffMember;
};

export function StaffActionsMenu({ member }: StaffActionsMenuProps) {
  const permissions = useSellerPermissions();
  const disableMutation = useDisableSellerStaff();
  const resendMutation = useResendSellerStaffInvite();
  const [error, setError] = useState("");

  const canResend =
    permissions.can("team:invite") && member.status === "invited";
  const canDisable =
    permissions.can("team:disable") &&
    member.status !== "disabled" &&
    member.role !== "seller";

  function resendInvite() {
    setError("");
    resendMutation.mutate(member.staff_id, {
      onError: (mutationError) => setError(getSafeErrorMessage(mutationError)),
    });
  }

  function disableStaff() {
    setError("");
    disableMutation.mutate(
      {
        staff_id: member.staff_id,
        status: "disabled",
      },
      {
        onError: (mutationError) => setError(getSafeErrorMessage(mutationError)),
      },
    );
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <div className="inline-flex items-center justify-end gap-2">
        {member.status === "invited" ? (
          <button
            type="button"
            disabled={!canResend || resendMutation.isPending}
            onClick={resendInvite}
            className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <Send className="h-3.5 w-3.5" aria-hidden="true" />
            {resendMutation.isPending ? "Sending" : "Resend"}
          </button>
        ) : null}

        <button
          type="button"
          disabled={!canDisable || disableMutation.isPending}
          onClick={disableStaff}
          className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-not-allowed disabled:opacity-50"
        >
          <UserX className="h-3.5 w-3.5" aria-hidden="true" />
          {disableMutation.isPending ? "Disabling" : "Disable"}
        </button>
      </div>

      {error ? (
        <p className="max-w-64 text-right text-xs text-rose-600" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
