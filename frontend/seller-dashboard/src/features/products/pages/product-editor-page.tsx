import { useQuery } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { Link, useNavigate, useParams } from "react-router-dom";

import { FailedState } from "../../../components/state/failed-state";
import { FormError } from "../../../components/state/form-error";
import { FormSkeleton } from "../../../components/state/loading-skeleton";
import { PermissionDeniedState } from "../../../components/state/permission-denied-state";
import { getRequestId, getSafeErrorMessage, isPermissionError } from "../../../lib/api-error";
import { useSellerStore } from "../../../stores/seller-store";
import { useSellerPermissions } from "../../team/hooks/use-seller-permissions";
import { getProduct } from "../api/seller-product-api";
import { ProductForm } from "../components/product-form";
import { ProductStatusBadge } from "../components/product-status-badge";
import { PublishProductButton } from "../components/publish-product-button";
import { productQueryKeys } from "../hooks/query-keys";
import { useCreateProduct, useUpdateProduct } from "../hooks/use-product-mutations";
import type { ProductInput } from "../types";

type ProductEditorPageProps = {
  mode: "create" | "edit";
};

export function ProductEditorPage({ mode }: ProductEditorPageProps) {
  const navigate = useNavigate();
  const { productId } = useParams();
  const activeSeller = useSellerStore((state) => state.activeSeller);
  const permissions = useSellerPermissions();
  const canWriteProducts = permissions.can("products:write");
  const createMutation = useCreateProduct();
  const updateMutation = useUpdateProduct(productId ?? "");

  const productQuery = useQuery({
    queryKey: productQueryKeys.detail(productId ?? ""),
    queryFn: () => getProduct(productId ?? ""),
    enabled: mode === "edit" && Boolean(productId) && canWriteProducts,
  });

  async function handleSubmit(input: ProductInput) {
    if (mode === "create") {
      await createMutation.mutateAsync(input);
      navigate("/seller/products");
      return;
    }

    if (productId) {
      await updateMutation.mutateAsync(input);
      navigate("/seller/products");
    }
  }

  const product = productQuery.data;
  const saveError = createMutation.error ?? updateMutation.error;
  const saving = createMutation.isPending || updateMutation.isPending;
  const sellerMismatch =
    mode === "edit" &&
    product?.seller_id &&
    activeSeller?.seller_id &&
    product.seller_id !== activeSeller.seller_id;

  if (!activeSeller) {
    return (
      <PermissionDeniedState
        title="Active seller unavailable"
        description="Product editor open karne ke liye active seller context required hai."
      />
    );
  }

  if (!canWriteProducts) {
    return (
      <PermissionDeniedState
        title="Product write permission required"
        description="Aapke current seller role ke paas products create ya update karne ka permission nahi hai."
      />
    );
  }

  if (mode === "edit" && !productId) {
    return (
      <FailedState
        title="Product id missing"
        description="Requested product route valid nahi hai."
        action={{ label: "Back to products", href: "/seller/products" }}
      />
    );
  }

  if (mode === "edit" && productQuery.isLoading) {
    return <FormSkeleton rows={6} />;
  }

  if (mode === "edit" && productQuery.isError) {
    if (isPermissionError(productQuery.error)) {
      return <PermissionDeniedState description={getSafeErrorMessage(productQuery.error)} />;
    }

    return (
      <FailedState
        title="Product could not be loaded"
        description={getSafeErrorMessage(productQuery.error)}
        requestId={getRequestId(productQuery.error)}
        action={{ label: "Retry", onClick: () => productQuery.refetch() }}
        secondaryAction={{
          label: "Back to products",
          href: "/seller/products",
          variant: "secondary",
        }}
      />
    );
  }

  if (sellerMismatch) {
    return (
      <PermissionDeniedState
        title="Product unavailable"
        description="This product is not available for the active seller."
      />
    );
  }

  return (
    <section className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <Link
            to="/seller/products"
            aria-label="Back to products"
            title="Back to products"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-slate-200 bg-white text-slate-600 transition hover:bg-slate-50 hover:text-slate-950 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          >
            <ArrowLeft className="h-4 w-4" aria-hidden="true" />
          </Link>
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              Product manager
            </p>
            <div className="mt-1 flex flex-wrap items-center gap-2">
              <h1 className="text-xl font-semibold text-slate-950">
                {mode === "create" ? "Create product" : product?.title || "Edit product"}
              </h1>
              {product ? <ProductStatusBadge status={product.status} /> : null}
            </div>
          </div>
        </div>

        {product ? <PublishProductButton product={product} /> : null}
      </div>

      {saveError ? (
        <FormError error={saveError} />
      ) : null}

      <ProductForm
        mode={mode}
        product={product}
        saving={saving}
        onSubmit={handleSubmit}
      />
    </section>
  );
}
