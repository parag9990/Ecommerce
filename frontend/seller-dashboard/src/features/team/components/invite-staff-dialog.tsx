import { zodResolver } from "@hookform/resolvers/zod";
import { X } from "lucide-react";
import { useEffect, useId, useState } from "react";
import { useForm } from "react-hook-form";

import { FormError } from "../../../components/state/form-error";
import { RoleSelect } from "./role-select";
import { useInviteSellerStaff } from "../hooks/use-seller-team";
import {
  emptyInviteStaffFormValues,
  inviteStaffSchema,
  type InviteStaffFormInput,
  type InviteStaffFormValues,
} from "../utils/team-validation";

type InviteStaffDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function InviteStaffDialog({ open, onOpenChange }: InviteStaffDialogProps) {
  const titleId = useId();
  const roleLabelId = useId();
  const inviteMutation = useInviteSellerStaff();
  const [submitError, setSubmitError] = useState<unknown>(null);

  const form = useForm<InviteStaffFormInput, unknown, InviteStaffFormValues>({
    resolver: zodResolver(inviteStaffSchema),
    defaultValues: emptyInviteStaffFormValues,
    mode: "onBlur",
  });

  useEffect(() => {
    if (!open) {
      form.reset(emptyInviteStaffFormValues);
      setSubmitError("");
    }
  }, [form, open]);

  async function onSubmit(values: InviteStaffFormValues) {
    try {
      setSubmitError("");
      await inviteMutation.mutateAsync(values);
      form.reset(emptyInviteStaffFormValues);
      onOpenChange(false);
    } catch (error) {
      setSubmitError(error);
    }
  }

  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        aria-label="Close invite dialog"
        className="absolute inset-0 bg-slate-950/40"
        onClick={() => onOpenChange(false)}
      />

      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="relative w-full max-w-lg rounded-md border border-slate-200 bg-white shadow-xl"
      >
        <div className="flex items-center justify-between gap-3 border-b border-slate-200 px-4 py-3">
          <div>
            <h2 id={titleId} className="text-base font-semibold text-slate-950">
              Invite staff
            </h2>
            <p className="mt-0.5 text-sm text-slate-500">
              Assign a scoped dashboard role before sending the invite.
            </p>
          </div>
          <button
            type="button"
            aria-label="Close"
            onClick={() => onOpenChange(false)}
            className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <X className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>

        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 p-4">
          <label className="space-y-1">
            <span className="text-sm font-medium text-slate-700">Email address</span>
            <input
              type="email"
              placeholder="name@example.com"
              {...form.register("email")}
              aria-invalid={Boolean(form.formState.errors.email)}
              className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            />
            {form.formState.errors.email?.message ? (
              <p className="text-xs text-rose-600">
                {form.formState.errors.email.message}
              </p>
            ) : null}
          </label>

          <div className="space-y-1">
            <label
              id={roleLabelId}
              htmlFor="invite-staff-role"
              className="block text-sm font-medium text-slate-700"
            >
              Role
            </label>
            <RoleSelect
              id="invite-staff-role"
              labelledBy={roleLabelId}
              value={form.watch("role")}
              disabled={inviteMutation.isPending}
              onChange={(role) =>
                form.setValue("role", role, {
                  shouldDirty: true,
                  shouldValidate: true,
                })
              }
            />
            {form.formState.errors.role?.message ? (
              <p className="text-xs text-rose-600">
                {form.formState.errors.role.message}
              </p>
            ) : null}
          </div>

          {submitError ? (
            <FormError error={submitError} />
          ) : null}

          <div className="flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="inline-flex h-10 items-center rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={inviteMutation.isPending}
              className="inline-flex h-10 items-center rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {inviteMutation.isPending ? "Sending invite..." : "Send invite"}
            </button>
          </div>
        </form>
      </section>
    </div>
  );
}
