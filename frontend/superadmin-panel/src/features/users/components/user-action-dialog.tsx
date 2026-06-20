import { useState } from "react";
import { Lock, Unlock } from "lucide-react";

import { ConfirmDialog } from "../../../components/ui/confirm-dialog";
import type { AdminUser, UserStatus } from "../types";
import { useUpdateUserStatus } from "../hooks/use-update-user-status";

const MIN_REASON_LENGTH = 10;

export function UserActionDialog({
  user,
  nextStatus
}: {
  user: AdminUser;
  nextStatus: Extract<UserStatus, "active" | "blocked">;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [reason, setReason] = useState("");
  const mutation = useUpdateUserStatus();
  const isBlocking = nextStatus === "blocked";
  const isReasonValid = reason.trim().length >= MIN_REASON_LENGTH;
  const Icon = isBlocking ? Lock : Unlock;

  const closeDialog = () => {
    if (mutation.isPending) {
      return;
    }

    setIsOpen(false);
    setReason("");
    mutation.reset();
  };

  const submit = async () => {
    if (!isReasonValid) {
      return;
    }

    await mutation.mutateAsync({
      userId: user.user_id,
      status: nextStatus,
      reason: reason.trim()
    });

    setReason("");
    setIsOpen(false);
  };

  return (
    <>
      <button
        type="button"
        onClick={() => setIsOpen(true)}
        className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-slate-300 bg-white px-3 text-sm font-medium text-slate-900 hover:bg-slate-50"
      >
        <Icon className="h-4 w-4" aria-hidden="true" />
        {isBlocking ? "Block user" : "Unblock user"}
      </button>

      <ConfirmDialog
        open={isOpen}
        title={isBlocking ? "Block user" : "Unblock user"}
        description="Add a clear audit reason before changing this user's account status."
        tone={isBlocking ? "danger" : "neutral"}
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
              disabled={!isReasonValid || mutation.isPending}
              onClick={submit}
              className="h-10 rounded-md bg-slate-950 px-3 text-sm font-medium text-white disabled:cursor-not-allowed disabled:bg-slate-400"
            >
              {mutation.isPending
                ? isBlocking
                  ? "Blocking..."
                  : "Unblocking..."
                : "Confirm"}
            </button>
          </>
        }
      >
        <label className="block">
          <span className="text-sm font-medium text-slate-700">Audit reason</span>
          <textarea
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            rows={4}
            className="mt-1 w-full resize-none rounded-lg border border-slate-300 p-3 text-sm outline-none focus:border-slate-900 focus:ring-2 focus:ring-slate-900/10"
            placeholder="Reason for this action"
          />
        </label>
        <div className="mt-2 text-xs text-slate-500">
          Minimum {MIN_REASON_LENGTH} characters. This reason is sent with the status update.
        </div>
        {mutation.error ? (
          <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {mutation.error instanceof Error ? mutation.error.message : "Status update failed."}
          </p>
        ) : null}
      </ConfirmDialog>
    </>
  );
}
