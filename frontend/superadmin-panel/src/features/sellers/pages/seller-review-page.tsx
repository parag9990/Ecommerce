import { ArrowLeft } from "lucide-react";
import { Link, useParams } from "react-router-dom";

import { DataState } from "../../../components/ui/data-state";
import { PermissionDenied } from "../../../components/ui/permission-denied";
import { CatalogReviewTable } from "../components/catalog-review-table";
import { KycDocumentList } from "../components/kyc-document-list";
import { SellerActionBar } from "../components/seller-action-bar";
import { SellerProfileSummary } from "../components/seller-profile-summary";
import { SellerReviewChecklist } from "../components/seller-review-checklist";
import { useAdminSeller } from "../hooks/use-admin-seller";
import { useSellerPermissions } from "../permissions";

export function SellerReviewPage() {
  const { sellerId = "" } = useParams();
  const permissions = useSellerPermissions();
  const sellerQuery = useAdminSeller(sellerId);

  if (!permissions.canViewSellers) {
    return <PermissionDenied compact />;
  }

  if (!sellerId) {
    return (
      <DataState
        tone="danger"
        title="Seller profile not found"
        description="The route did not include a valid seller id."
      />
    );
  }

  if (sellerQuery.isLoading) {
    return <DataState title="Loading seller" description="Fetching the selected seller profile." />;
  }

  if (sellerQuery.error) {
    return (
      <DataState
        tone="danger"
        title="Unable to load seller"
        description={
          sellerQuery.error instanceof Error ? sellerQuery.error.message : "The seller profile request failed."
        }
        action={
          <button
            type="button"
            onClick={() => void sellerQuery.refetch()}
            className="h-9 rounded-md border border-red-200 bg-white px-3 font-medium text-red-800 hover:bg-red-100"
          >
            Retry
          </button>
        }
      />
    );
  }

  if (!sellerQuery.data) {
    return <DataState tone="danger" title="Seller profile not found" description="No seller matched this id." />;
  }

  const seller = sellerQuery.data;

  return (
    <section className="min-h-[calc(100vh-6.5rem)] overflow-hidden rounded-lg border border-slate-200 bg-slate-50">
      <header className="border-b border-slate-200 bg-white px-4 py-4">
        <Link
          to="/admin/sellers"
          className="mb-3 inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-950"
        >
          <ArrowLeft size={16} aria-hidden="true" />
          Back to sellers
        </Link>
        <div className="min-w-0">
          <h1 className="truncate text-xl font-semibold text-slate-950">{seller.store_name}</h1>
          <p className="mt-1 break-all text-sm text-slate-600">{seller.seller_id}</p>
        </div>
      </header>

      <div className="grid gap-4 p-4 xl:grid-cols-[minmax(0,1fr)_360px]">
        <main className="min-w-0 space-y-4">
          <SellerProfileSummary seller={seller} canViewKyc={permissions.canReviewSellerKyc} />
          <KycDocumentList
            sellerId={seller.seller_id}
            canViewKyc={permissions.canReviewSellerKyc}
            canViewMaskedKyc={permissions.canViewMaskedSellerKyc}
          />
          <CatalogReviewTable sellerId={seller.seller_id} enabled={permissions.canViewSellerCatalog} />
        </main>

        <div className="space-y-4">
          <SellerActionBar seller={seller} roles={permissions.roles} />
          <SellerReviewChecklist seller={seller} />
        </div>
      </div>
    </section>
  );
}
