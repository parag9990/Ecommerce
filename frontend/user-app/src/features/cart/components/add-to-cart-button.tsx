import { Check, ShoppingCart } from 'lucide-react';
import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';

import { Button } from '../../../components/ui/button';
import { ApiError } from '../../../lib/http';
import { routePaths } from '../../../routes/route-paths';
import { useCartMutations } from '../hooks/use-cart-mutations';

type AddToCartButtonProps = {
  compact?: boolean | undefined;
  disabled?: boolean | undefined;
  fullWidth?: boolean | undefined;
  productId: string;
  quantity?: number | undefined;
  showViewCartLink?: boolean | undefined;
  variantId?: string | undefined;
};

type AddToCartStatus = 'idle' | 'loading' | 'success' | 'error';

export function AddToCartButton({
  compact = false,
  disabled = false,
  fullWidth = false,
  productId,
  quantity = 1,
  showViewCartLink = true,
  variantId,
}: AddToCartButtonProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [status, setStatus] = useState<AddToCartStatus>('idle');
  const [error, setError] = useState<string>();
  const { addItem } = useCartMutations();
  const isDisabled = disabled || !variantId || addItem.isPending;

  async function handleAddToCart() {
    if (!variantId) {
      return;
    }

    setError(undefined);
    setStatus('loading');

    try {
      await addItem.mutateAsync({
        product_id: productId,
        quantity,
        variant_id: variantId,
      });
      setStatus('success');
    } catch (apiError) {
      if (apiError instanceof ApiError && apiError.status === 401) {
        const redirectTo = encodeURIComponent(location.pathname + location.search);
        void navigate(`${routePaths.login}?redirect=${redirectTo}`);
        return;
      }

      setStatus('error');
      setError(
        apiError instanceof Error
          ? apiError.message
          : 'Item could not be added to cart.',
      );
    }
  }

  return (
    <div className={fullWidth ? 'w-full' : ''}>
      <Button
        className={[
          compact ? 'h-9 px-3' : '',
          fullWidth ? 'w-full' : '',
          status === 'success' ? 'bg-emerald-600 hover:bg-emerald-700' : '',
        ].join(' ')}
        disabled={isDisabled}
        onClick={() => {
          void handleAddToCart();
        }}
        type="button"
      >
        {status === 'success' ? (
          <Check aria-hidden="true" className="mr-2 h-4 w-4" />
        ) : (
          <ShoppingCart aria-hidden="true" className="mr-2 h-4 w-4" />
        )}
        {addItem.isPending ? 'Adding' : compact ? 'Add' : 'Add to cart'}
      </Button>

      {status === 'success' && showViewCartLink ? (
        <Link
          className="mt-2 inline-block rounded text-sm font-semibold text-blue-700 hover:text-blue-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
          to={routePaths.cart}
        >
          View cart
        </Link>
      ) : null}

      {status === 'error' && error ? (
        <p className="mt-2 text-sm text-red-700">{error}</p>
      ) : null}
    </div>
  );
}
