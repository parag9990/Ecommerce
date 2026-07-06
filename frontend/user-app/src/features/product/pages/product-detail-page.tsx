import { useMemo, useState } from 'react';
import { Link, useParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { AddToCartButton } from '../../cart/components/add-to-cart-button';
import { WishlistButton } from '../../wishlist/components/wishlist-button';
import { Price } from '../components/price';
import { ProductDetailSkeleton } from '../components/product-detail-skeleton';
import { ProductImageGallery } from '../components/product-image-gallery';
import { VariantPicker } from '../components/variant-picker';
import { useProductDetail } from '../hooks/use-product-detail';

function DetailRow({
  label,
  value,
}: {
  label: string;
  value?: string | undefined;
}) {
  if (!value) {
    return null;
  }

  return (
    <div className="flex justify-between gap-4 border-b border-slate-100 py-2 last:border-b-0">
      <dt className="text-slate-500">{label}</dt>
      <dd className="text-right font-medium text-slate-800">{value}</dd>
    </div>
  );
}

export function ProductDetailPage() {
  const { productId } = useParams();
  const { error, isLoading, product } = useProductDetail(productId);
  const variants = useMemo(() => product?.variants ?? [], [product?.variants]);
  const [selection, setSelection] = useState<{
    productId?: string | undefined;
    variantId?: string | undefined;
  }>({});
  const selectedVariantId =
    selection.productId === product?.product_id ? selection.variantId : undefined;

  const selectedVariant =
    variants.find(
      (variant) => (variant.variant_id ?? variant.sku) === selectedVariantId,
    ) ?? variants[0];
  const selectedStock =
    selectedVariant?.available_quantity ?? selectedVariant?.stock_quantity;
  const selectedCartVariantId = selectedVariant?.variant_id ?? selectedVariant?.sku;

  if (isLoading) {
    return <ProductDetailSkeleton />;
  }

  if (error) {
    return (
      <div className="mx-auto max-w-3xl">
        <Alert title="Product details could not be loaded" variant="error">
          {error}
        </Alert>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="mx-auto max-w-3xl">
        <EmptyState
          action={
            <Link
              className="text-sm font-semibold text-blue-700 underline-offset-4 hover:text-blue-800 hover:underline"
              to={routePaths.home}
            >
              Back to home
            </Link>
          }
          description="This product may have been removed, unpublished, or opened from an invalid link."
          title="Product not found"
        />
      </div>
    );
  }

  return (
    <div className="grid gap-8 lg:grid-cols-2">
      <ProductImageGallery images={product.images} title={product.title} />

      <section className="space-y-6">
        <div className="space-y-2">
          {product.brand ? (
            <p className="text-sm font-medium uppercase tracking-wide text-slate-500">
              {product.brand}
            </p>
          ) : null}
          <h1 className="text-3xl font-semibold tracking-tight text-slate-950">
            {product.title}
          </h1>
          <div className="text-xl">
            <Price value={selectedVariant?.price} />
          </div>
        </div>

        {product.description ? (
          <p className="text-sm leading-6 text-slate-700">
            {product.description}
          </p>
        ) : (
          <p className="text-sm leading-6 text-slate-500">
            Product description is not available yet.
          </p>
        )}

        <VariantPicker
          onChange={(variantId) => {
            setSelection({
              productId: product.product_id,
              variantId,
            });
          }}
          selectedVariantId={selectedVariant ? selectedVariant.variant_id ?? selectedVariant.sku : undefined}
          variants={variants}
        />

        <div className="grid gap-3 rounded-md border border-slate-200 bg-white p-4 sm:grid-cols-[minmax(0,1fr)_auto]">
          {selectedVariant ? (
            <AddToCartButton
              disabled={selectedStock === 0}
              fullWidth
              productId={product.product_id}
              variantId={selectedCartVariantId}
            />
          ) : (
            <p className="text-sm text-slate-600">
              Select a variant to add this product to your cart.
            </p>
          )}
          <WishlistButton
            productId={product.product_id}
            variantId={selectedCartVariantId}
          />
        </div>

        <dl className="rounded-md border border-slate-200 bg-white p-4 text-sm">
          <DetailRow label="Product ID" value={product.product_id} />
          <DetailRow label="Seller" value={product.seller_id} />
          <DetailRow label="Category" value={product.category_id} />
          <DetailRow label="Status" value={product.status} />
          <DetailRow label="Selected SKU" value={selectedVariant?.sku} />
        </dl>
      </section>
    </div>
  );
}
