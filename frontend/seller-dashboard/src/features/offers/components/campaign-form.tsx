import { zodResolver } from "@hookform/resolvers/zod";
import { CalendarPlus } from "lucide-react";
import { useForm } from "react-hook-form";

import type { CampaignInput } from "../types";
import { toCampaignInput } from "../utils/offer-mappers";
import type { CampaignFormInput, CampaignFormValues } from "../utils/offer-validation";
import { campaignFormSchema, emptyCampaignFormValues } from "../utils/offer-validation";

type CampaignFormProps = {
  saving: boolean;
  onSubmit: (input: CampaignInput) => Promise<void>;
};

function FormError({ message }: { message?: string }) {
  if (!message) {
    return null;
  }

  return <p className="text-xs text-rose-600">{message}</p>;
}

export function CampaignForm({ saving, onSubmit }: CampaignFormProps) {
  const form = useForm<CampaignFormInput, unknown, CampaignFormValues>({
    resolver: zodResolver(campaignFormSchema),
    defaultValues: emptyCampaignFormValues,
    mode: "onBlur",
  });

  return (
    <form
      onSubmit={form.handleSubmit((values) => onSubmit(toCampaignInput(values)))}
      className="space-y-4 rounded-md border border-slate-200 bg-white p-4 shadow-sm"
    >
      <label className="space-y-1">
        <span className="text-sm font-medium text-slate-700">Campaign name</span>
        <input
          {...form.register("name")}
          className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          placeholder="Festive Sale"
          aria-invalid={Boolean(form.formState.errors.name)}
        />
        <FormError message={form.formState.errors.name?.message} />
      </label>

      <div className="grid gap-4 md:grid-cols-2">
        <label className="space-y-1">
          <span className="text-sm font-medium text-slate-700">Starts at</span>
          <input
            type="datetime-local"
            {...form.register("starts_at")}
            className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            aria-invalid={Boolean(form.formState.errors.starts_at)}
          />
          <FormError message={form.formState.errors.starts_at?.message} />
        </label>

        <label className="space-y-1">
          <span className="text-sm font-medium text-slate-700">Ends at</span>
          <input
            type="datetime-local"
            {...form.register("ends_at")}
            className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            aria-invalid={Boolean(form.formState.errors.ends_at)}
          />
          <FormError message={form.formState.errors.ends_at?.message} />
        </label>
      </div>

      <label className="max-w-sm space-y-1">
        <span className="text-sm font-medium text-slate-700">Budget amount</span>
        <input
          type="number"
          min="0"
          {...form.register("budget.amount")}
          className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          placeholder="Optional"
        />
        <FormError message={form.formState.errors.budget?.amount?.message} />
      </label>

      <input type="hidden" {...form.register("budget.currency")} value="INR" />

      <div className="flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
        <button
          type="submit"
          disabled={saving}
          className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-4 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <CalendarPlus className="h-4 w-4" aria-hidden="true" />
          {saving ? "Creating..." : "Create campaign"}
        </button>
      </div>
    </form>
  );
}
