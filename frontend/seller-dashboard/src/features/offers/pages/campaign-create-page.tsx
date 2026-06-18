import { ArrowLeft } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";

import { FormError } from "../../../components/state/form-error";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { CampaignForm } from "../components/campaign-form";
import { useCreateCampaign } from "../hooks/use-campaign-mutations";
import type { CampaignInput } from "../types";

export function CampaignCreatePage() {
  const navigate = useNavigate();
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canWriteOffers = permissions.can("offers:write");
  const createCampaign = useCreateCampaign();

  async function handleSubmit(input: CampaignInput) {
    await createCampaign.mutateAsync(input);
    navigate("/seller/offers?tab=campaigns");
  }

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Campaign create karne ke liye active seller context required hai."
      />
    );
  }

  if (!canWriteOffers) {
    return (
      <PermissionDeniedState
        title="Offers write permission required"
        description="Aapke current seller role ke paas campaigns create karne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <Link
            to="/seller/offers?tab=campaigns"
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
            <h1 className="mt-1 text-xl font-semibold text-slate-950">Create campaign</h1>
          </div>
        </div>
      </div>

      {createCampaign.error ? (
        <FormError error={createCampaign.error} />
      ) : null}

      <CampaignForm saving={createCampaign.isPending} onSubmit={handleSubmit} />
    </section>
  );
}
