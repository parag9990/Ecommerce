import { Link } from 'react-router-dom';
import { useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { EmptyState } from '../../../components/ui/empty-state';
import { routePaths } from '../../../routes/route-paths';
import { WishlistGrid } from '../components/wishlist-grid';
import { useWishlistMutations } from '../hooks/use-wishlist-mutations';
import { useWishlistQuery } from '../hooks/use-wishlist-query';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function WishlistSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: 4 }, (_, index) => (
        <div
          className="h-80 animate-pulse rounded-md border border-slate-200 bg-white"
          key={index}
        />
      ))}
    </div>
  );
}

export function WishlistPage() {
  const [busyProductId, setBusyProductId] = useState<string>();
  const [actionError, setActionError] = useState<string>();
  const [success, setSuccess] = useState<string>();
  const wishlistQuery = useWishlistQuery();
  const { moveToCart: moveToCartMutation, removeItem: removeItemMutation } =
    useWishlistMutations();

  async function removeItem(productId: string) {
    setBusyProductId(productId);
    setActionError(undefined);
    setSuccess(undefined);

    try {
      await removeItemMutation.mutateAsync(productId);
      setSuccess('Item removed from wishlist.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Item could not be removed.'));
    } finally {
      setBusyProductId(undefined);
    }
  }

  async function moveToCart(productId: string) {
    setBusyProductId(productId);
    setActionError(undefined);
    setSuccess(undefined);

    try {
      await moveToCartMutation.mutateAsync(productId);
      setSuccess('Item moved to cart.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Item could not be moved to cart.'));
    } finally {
      setBusyProductId(undefined);
    }
  }

  const error =
    wishlistQuery.error instanceof Error
      ? wishlistQuery.error.message
      : wishlistQuery.isError
        ? 'Wishlist could not be loaded.'
        : undefined;
  const items = wishlistQuery.data?.items ?? [];
  const isEmpty = !wishlistQuery.isLoading && items.length === 0 && !error;

  return (
    <section className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Wishlist
        </h1>
        <p className="mt-1 text-sm text-slate-600">
          Save products for later and move them to your cart when ready.
        </p>
      </div>

      {error ? <Alert variant="error">{error}</Alert> : null}
      {actionError ? <Alert variant="error">{actionError}</Alert> : null}
      {success ? <Alert variant="success">{success}</Alert> : null}

      {wishlistQuery.isLoading ? <WishlistSkeleton /> : null}

      {isEmpty ? (
        <EmptyState
          action={
            <Link
              className="inline-flex h-10 items-center justify-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              to={routePaths.home}
            >
              Browse products
            </Link>
          }
          description="Products you save from listing or detail pages will show up here."
          title="Your wishlist is empty"
        />
      ) : null}

      {!wishlistQuery.isLoading && items.length > 0 ? (
        <WishlistGrid
          busyProductId={busyProductId}
          items={items}
          onMoveToCart={(productId) => {
            void moveToCart(productId);
          }}
          onRemove={(productId) => {
            void removeItem(productId);
          }}
        />
      ) : null}
    </section>
  );
}
