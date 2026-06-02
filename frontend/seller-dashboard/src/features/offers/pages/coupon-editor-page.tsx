import { ArrowLeft } from "lucide-react";
import { useMemo } from "react";
import { Link, useLocation, useNavigate, useParams } from "react-router-dom";

import { FailedState } from "../../../components/state/failed-state";
import { FormError } from "../../../components/state/form-error";
import { FormSkeleton } from "../../../components/state/loading-skeleton";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { getRequestId, getSafeErrorMessage, isPermissionError } from "../../../lib/api-error";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { CouponForm } from "../components/coupon-form";
import { CouponStatusBadge } from "../components/coupon-status-badge";
import { useCreateCoupon, useUpdateCoupon } from "../hooks/use-coupon-mutations";
import { useSellerCoupons } from "../hooks/use-seller-coupons";
import type { Coupon, CouponInput } from "../types";
import { toCouponFormValues } from "../utils/offer-mappers";

type CouponEditorPageProps = {
  mode: "create" | "edit";
};

export function CouponEditorPage({ mode }: CouponEditorPageProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const { couponId } = useParams();
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canWriteOffers = permissions.can("offers:write");
  const routedCoupon = (location.state as { coupon?: Coupon } | null)?.coupon;
  const createMutation = useCreateCoupon();
  const updateMutation = useUpdateCoupon(couponId ?? "");
  const couponsQuery = useSellerCoupons(
    { status: "all", page: 1, page_size: 100 },
    mode === "edit" && canWriteOffers ? activeSeller?.seller_id : undefined,
  );

  const coupon = useMemo(() => {
    const freshCoupon = couponsQuery.data?.coupons.find(
      (item) => item.coupon_id === couponId,
    );

    return freshCoupon ?? routedCoupon;
  }, [couponId, couponsQuery.data?.coupons, routedCoupon]);

  async function handleSubmit(input: CouponInput) {
    if (mode === "create") {
      await createMutation.mutateAsync(input);
      navigate("/seller/offers?tab=coupons");
      return;
    }

    if (couponId) {
      await updateMutation.mutateAsync(input);
      navigate("/seller/offers?tab=coupons");
    }
  }

  const saveError = createMutation.error ?? updateMutation.error;
  const saving = createMutation.isPending || updateMutation.isPending;
  const formValues = useMemo(() => toCouponFormValues(coupon), [coupon]);

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Coupon editor open karne ke liye active seller context required hai."
      />
    );
  }

  if (!canWriteOffers) {
    return (
      <PermissionDeniedState
        title="Offers write permission required"
        description="Aapke current seller role ke paas coupons create ya update karne ka permission nahi hai."
      />
    );
  }

  if (mode === "edit" && !couponId) {
    return (
      <FailedState
        title="Coupon id missing"
        description="Requested coupon route valid nahi hai."
        action={{ label: "Back to offers", href: "/seller/offers?tab=coupons" }}
      />
    );
  }

  if (mode === "edit" && couponsQuery.isLoading && !routedCoupon) {
    return <FormSkeleton rows={5} />;
  }

  if (mode === "edit" && couponsQuery.isError && !routedCoupon) {
    if (isPermissionError(couponsQuery.error)) {
      return <PermissionDeniedState description={getSafeErrorMessage(couponsQuery.error)} />;
    }

    return (
      <FailedState
        title="Coupon could not be loaded"
        description={getSafeErrorMessage(couponsQuery.error)}
        requestId={getRequestId(couponsQuery.error)}
        action={{ label: "Retry", onClick: () => couponsQuery.refetch() }}
        secondaryAction={{
          label: "Back to offers",
          href: "/seller/offers?tab=coupons",
          variant: "secondary",
        }}
      />
    );
  }

  if (mode === "edit" && couponsQuery.data && !coupon) {
    return (
      <FailedState
        title="Coupon not found"
        description="Coupon current seller coupon list me nahi mila."
        action={{ label: "Back to offers", href: "/seller/offers?tab=coupons" }}
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <Link
            to="/seller/offers?tab=coupons"
            aria-label="Back to offers"
            title="Back to offers"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          </Link>
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Offers
            </p>
            <div className="mt-1 flex flex-wrap items-center gap-2">
              <h1 className="text-xl font-semibold text-slate-950">
                {mode === "create" ? "Create coupon" : coupon?.code || "Edit coupon"}
              </h1>
              {coupon ? <CouponStatusBadge status={coupon.status} /> : null}
            </div>
          </div>
        </div>
      </div>

      {saveError ? (
        <FormError error={saveError} />
      ) : null}

      <CouponForm
        defaultValues={formValues}
        saving={saving}
        onSubmit={handleSubmit}
      />
    </section>
  );
}
