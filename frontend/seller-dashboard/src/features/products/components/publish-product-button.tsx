import { Send } from "lucide-react";

import { getRequestId, getSafeErrorMessage } from "../../../lib/api-error";
import { usePublishProduct } from "../hooks/use-product-mutations";
import type { Product } from "../types";

type PublishProductButtonProps = {
  product: Product;
};

export function PublishProductButton({ product }: PublishProductButtonProps) {
  const publishMutation = usePublishProduct();
  const canPublish = ["draft", "approved", "unpublished"].includes(product.status);

  if (!canPublish) {
    return null;
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <button
        type="button"
        disabled={publishMutation.isPending}
        onClick={() => publishMutation.mutate(product.product_id)}
        className="inline-flex h-10 items-center gap-2 rounded-md bg-emerald-600 px-3 text-sm font-medium text-white transition hover:bg-emerald-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-700 disabled:opacity-60"
      >
        <Send className="h-4 w-4" aria-hidden="true" />
        {publishMutation.isPending ? "Publishing..." : "Publish"}
      </button>
      {publishMutation.isError ? (
        <p className="max-w-64 text-right text-xs text-rose-600" role="alert">
          {getSafeErrorMessage(publishMutation.error)}
          {getRequestId(publishMutation.error)
            ? ` Request ID: ${getRequestId(publishMutation.error)}`
            : ""}
        </p>
      ) : null}
    </div>
  );
}
