import { zodResolver } from "@hookform/resolvers/zod";
import { Save } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";

import type { CouponInput } from "../types";
import { toCouponFormValues, toCouponInput } from "../utils/offer-mappers";
import type { CouponFormInput, CouponFormValues } from "../utils/offer-validation";
import { couponFormSchema, emptyCouponFormValues } from "../utils/offer-validation";
import { DiscountTypeSelector } from "./discount-type-selector";

type CouponFormProps = {
  defaultValues?: CouponFormInput;
  saving: boolean;
  onSubmit: (input: CouponInput) => Promise<void>;
};

function FormError({ message }: { message?: string }) {
  if (!message) {
    return null;
  }

  return <p className="text-xs text-rose-600">{message}</p>;
}

export function CouponForm({ defaultValues, saving, onSubmit }: CouponFormProps) {
  const form = useForm<CouponFormInput, unknown, CouponFormValues>({
    resolver: zodResolver(couponFormSchema),
    defaultValues: defaultValues ?? emptyCouponFormValues,
    mode: "onBlur",
  });
  const discountType = form.watch("discount_type");
  const codeField = form.register("code");
  const { reset } = form;

  useEffect(() => {
    reset(defaultValues ?? toCouponFormValues());
  }, [defaultValues, reset]);

  return (
    <form
      onSubmit={form.handleSubmit((values) => onSubmit(toCouponInput(values)))}
      className="space-y-4 rounded-md border border-slate-200 bg-white p-4 shadow-sm"
    >
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Coupon code</span>
              <input
                {...codeField}
                onChange={(event) => {
                  event.target.value = event.target.value.toUpperCase();
                  codeField.onChange(event);
                }}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm uppercase outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="SAVE10"
                aria-invalid={Boolean(form.formState.errors.code)}
              />
              <FormError message={form.formState.errors.code?.message} />
            </label>

            <div className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Discount type</span>
              <DiscountTypeSelector
                value={discountType}
                onChange={(value) =>
                  form.setValue("discount_type", value, {
                    shouldDirty: true,
                    shouldValidate: true,
                  })
                }
              />
            </div>
          </div>

          <div className="grid gap-4 md:grid-cols-3">
            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">
                {discountType === "percentage" ? "Discount percent" : "Discount amount"}
              </span>
              <input
                type="number"
                min="1"
                max={discountType === "percentage" ? 100 : undefined}
                {...form.register("discount_value")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                aria-invalid={Boolean(form.formState.errors.discount_value)}
              />
              <FormError message={form.formState.errors.discount_value?.message} />
            </label>

            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Minimum cart amount</span>
              <input
                type="number"
                min="0"
                {...form.register("min_cart_amount.amount")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="0"
              />
              <FormError message={form.formState.errors.min_cart_amount?.amount?.message} />
            </label>

            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Usage limit</span>
              <input
                type="number"
                min="1"
                {...form.register("usage_limit")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="No limit"
              />
              <FormError message={form.formState.errors.usage_limit?.message} />
            </label>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Starts at</span>
              <input
                type="datetime-local"
                {...form.register("starts_at")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
              <FormError message={form.formState.errors.starts_at?.message} />
            </label>

            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Ends at</span>
              <input
                type="datetime-local"
                {...form.register("ends_at")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
              <FormError message={form.formState.errors.ends_at?.message} />
            </label>
          </div>

          <input type="hidden" {...form.register("min_cart_amount.currency")} value="INR" />
        </div>

        <aside className="rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800">
          <h2 className="font-semibold">API boundary</h2>
          <p className="mt-1">
            Product/category scope, max cap, new-user rules, and per-user limits are not
            part of the current coupon request schema.
          </p>
        </aside>
      </div>

      <div className="flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
        <button
          type="submit"
          disabled={saving}
          className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-4 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <Save className="h-4 w-4" aria-hidden="true" />
          {saving ? "Saving..." : "Save coupon"}
        </button>
      </div>
    </form>
  );
}
