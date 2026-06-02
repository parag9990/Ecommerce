import { CalendarPlus, Plus } from "lucide-react";
import { useState } from "react";
import { Link, useSearchParams } from "react-router-dom";

import { AsyncStateBoundary } from "../../../components/state/async-state-boundary";
import { EmptyState } from "../../../components/state/empty-state";
import { CardsSkeleton, TableSkeleton } from "../../../components/state/loading-skeleton";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { RefreshingNotice } from "../../../components/state/refreshing-notice";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { CampaignCalendar } from "../components/campaign-calendar";
import { CouponFiltersBar } from "../components/coupon-filters-bar";
import { CouponTable } from "../components/coupon-table";
import { OffersTabs } from "../components/offers-tabs";
import { UsageStatsCards } from "../components/usage-stats-cards";
import { useSellerCampaigns } from "../hooks/use-seller-campaigns";
import { useSellerCoupons } from "../hooks/use-seller-coupons";
import type { CouponFilters, OffersTab } from "../types";
import { formatMonthParam } from "../utils/offer-date-rules";

const defaultCouponFilters: CouponFilters = {
  status: "all",
  page: 1,
  page_size: 20,
};

function getActiveTab(value: string | null): OffersTab {
  return value === "campaigns" ? "campaigns" : "coupons";
}

export function OffersPage() {
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canViewOffers = permissions.can("offers:view");
  const canWriteOffers = permissions.can("offers:write");
  const [searchParams, setSearchParams] = useSearchParams();
  const [couponFilters, setCouponFilters] = useState<CouponFilters>(defaultCouponFilters);
  const [campaignMonth, setCampaignMonth] = useState(() => new Date());
  const activeTab = getActiveTab(searchParams.get("tab"));

  const couponsQuery = useSellerCoupons(
    couponFilters,
    canViewOffers ? activeSeller?.seller_id : undefined,
  );
  const campaignsQuery = useSellerCampaigns(
    {
      status: "all",
      month: formatMonthParam(campaignMonth),
      page: 1,
      page_size: 100,
    },
    canViewOffers ? activeSeller?.seller_id : undefined,
  );

  function setActiveTab(tab: OffersTab) {
    setSearchParams((current) => {
      const next = new URLSearchParams(current);
      next.set("tab", tab);
      return next;
    });
  }

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Offers load karne ke liye active seller context required hai."
      />
    );
  }

  if (!canViewOffers) {
    return (
      <PermissionDeniedState
        title="Offers access unavailable"
        description="Aapke current seller role ke paas offers and coupons dekhne ka permission nahi hai."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Promotions
          </p>
          <h1 className="mt-1 text-xl font-semibold text-slate-950">
            Offers and coupons
          </h1>
        </div>

        {canWriteOffers ? (
          <div className="flex flex-wrap items-center gap-2">
            <Link
              to="/seller/offers/campaigns/new"
              className="inline-flex h-10 items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            >
              <CalendarPlus className="h-4 w-4" aria-hidden="true" />
              New campaign
            </Link>
            <Link
              to="/seller/offers/coupons/new"
              className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-3 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950"
            >
              <Plus className="h-4 w-4" aria-hidden="true" />
              New coupon
            </Link>
          </div>
        ) : null}
      </div>

      {couponsQuery.isPending && campaignsQuery.isPending ? (
        <CardsSkeleton count={5} className="xl:grid-cols-5" />
      ) : (
        <UsageStatsCards
          coupons={couponsQuery.data?.coupons}
          campaigns={campaignsQuery.data?.campaigns}
        />
      )}

      <OffersTabs activeTab={activeTab} onChange={setActiveTab} />

      {activeTab === "coupons" ? (
        <div className="space-y-3">
          <CouponFiltersBar filters={couponFilters} onChange={setCouponFilters} />

          <RefreshingNotice
            show={couponsQuery.isFetching && !couponsQuery.isPending}
          />

          <AsyncStateBoundary
            isLoading={couponsQuery.isPending}
            isError={couponsQuery.isError}
            error={couponsQuery.error}
            data={couponsQuery.data}
            isEmpty={(data) => data.coupons.length === 0}
            loadingFallback={<TableSkeleton columns={7} />}
            emptyFallback={
              <EmptyState
                title="No coupons yet"
                description="First coupon create karke offers start karein."
                action={
                  canWriteOffers
                    ? { label: "Create coupon", href: "/seller/offers/coupons/new" }
                    : undefined
                }
              />
            }
            onRetry={() => couponsQuery.refetch()}
          >
            {(data) => (
              <CouponTable
                coupons={data.coupons}
                total={data.total}
                filters={couponFilters}
                onFiltersChange={setCouponFilters}
              />
            )}
          </AsyncStateBoundary>
        </div>
      ) : null}

      {activeTab === "campaigns" ? (
        <div className="space-y-3">
          <RefreshingNotice
            show={campaignsQuery.isFetching && !campaignsQuery.isPending}
          />

          <AsyncStateBoundary
            isLoading={campaignsQuery.isPending}
            isError={campaignsQuery.isError}
            error={campaignsQuery.error}
            data={campaignsQuery.data}
            isEmpty={(data) => data.campaigns.length === 0}
            loadingFallback={<TableSkeleton columns={7} rows={5} />}
            emptyFallback={
              <EmptyState
                title="No campaigns scheduled"
                description="Campaign calendar abhi empty hai."
                action={
                  canWriteOffers
                    ? { label: "Create campaign", href: "/seller/offers/campaigns/new" }
                    : undefined
                }
              />
            }
            onRetry={() => campaignsQuery.refetch()}
          >
            {(data) => (
              <CampaignCalendar
                campaigns={data.campaigns}
                month={campaignMonth}
                onMonthChange={setCampaignMonth}
              />
            )}
          </AsyncStateBoundary>
        </div>
      ) : null}
    </section>
  );
}
