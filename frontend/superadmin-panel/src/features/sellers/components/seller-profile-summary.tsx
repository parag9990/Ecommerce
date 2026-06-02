import { Mail, Store } from "lucide-react";

import { StatusBadge } from "../../../components/ui/status-badge";
import { formatDateTime } from "../../../lib/format";
import type { AdminSeller } from "../types";

function fieldValue(value?: string | null): string {
  return value?.trim() || "Not added";
}

function businessId(value: string | null | undefined, canViewKyc: boolean): string {
  if (!value) {
    return "Not added";
  }

  return canViewKyc ? value : "Masked";
}

export function SellerProfileSummary({
  seller,
  canViewKyc
}: {
  seller: AdminSeller;
  canViewKyc: boolean;
}) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-slate-600">
            <Store className="h-4 w-4" aria-hidden="true" />
            <span className="text-sm font-medium">Seller profile</span>
          </div>
          <h2 className="mt-2 truncate text-lg font-semibold text-slate-950">{seller.store_name}</h2>
          <p className="mt-1 break-all text-sm text-slate-600">{seller.seller_id}</p>
        </div>
        <StatusBadge status={seller.status} />
      </div>

      <dl className="mt-4 grid gap-3 text-sm sm:grid-cols-2 xl:grid-cols-3">
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Owner</dt>
          <dd className="text-slate-950">{fieldValue(seller.display_name)}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Support Email</dt>
          <dd className="flex min-w-0 items-center gap-2 text-slate-950">
            <Mail className="h-4 w-4 shrink-0 text-slate-500" aria-hidden="true" />
            <span className="truncate">{fieldValue(seller.support_email)}</span>
          </dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">GST Number</dt>
          <dd className="break-all text-slate-950">{businessId(seller.gst_number, canViewKyc)}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">KYC Status</dt>
          <dd className="mt-1">
            <StatusBadge status={seller.kyc_status ?? "not_started"} />
          </dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Created</dt>
          <dd className="text-slate-950">{formatDateTime(seller.created_at)}</dd>
        </div>
        <div>
          <dt className="text-xs font-semibold uppercase text-slate-500">Approved</dt>
          <dd className="text-slate-950">{formatDateTime(seller.approved_at)}</dd>
        </div>
      </dl>
    </section>
  );
}
