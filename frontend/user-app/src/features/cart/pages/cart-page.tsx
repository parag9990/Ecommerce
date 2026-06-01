import { Alert } from '../../../components/ui/alert';
import { CartEmptyState } from '../components/cart-empty-state';
import { CartItemRow } from '../components/cart-item-row';
import { CartSummary } from '../components/cart-summary';
import { CouponBox } from '../components/coupon-box';
import {
  clearCartCouponCode,
  readCartCouponCode,
  storeCartCouponCode,
} from '../cart-coupon';
import { useCart } from '../hooks/use-cart';

function CartLoadingState() {
  return (
    <div className="space-y-4">
      {Array.from({ length: 3 }, (_, index) => (
        <div
          className="grid animate-pulse gap-4 rounded-md border border-slate-200 bg-white p-4 sm:grid-cols-[6rem_1fr_8rem]"
          key={index}
        >
          <div className="aspect-square rounded-md bg-slate-200" />
          <div className="space-y-3">
            <div className="h-4 w-2/3 rounded bg-slate-200" />
            <div className="h-3 w-1/3 rounded bg-slate-200" />
          </div>
          <div className="h-10 rounded bg-slate-200" />
        </div>
      ))}
    </div>
  );
}

export function CartPage() {
  const {
    actionError,
    cart,
    changeQuantity,
    error,
    isLoading,
    mutatingItemId,
    removeItem,
  } = useCart();
  const items = cart?.items ?? [];
  const hasOutOfStockItem = items.some(
    (item) => item.stock_status === 'out_of_stock',
  );
  const isEmpty = !isLoading && items.length === 0;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Cart
        </h1>
        <CartLoadingState />
      </div>
    );
  }

  if (error) {
    return (
      <div className="mx-auto max-w-3xl">
        <Alert title="Cart could not be loaded" variant="error">
          {error}
        </Alert>
      </div>
    );
  }

  if (isEmpty) {
    return <CartEmptyState />;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Cart
        </h1>
      </div>

      {actionError ? <Alert variant="error">{actionError}</Alert> : null}

      {hasOutOfStockItem ? (
        <Alert variant="error">
          Remove out-of-stock items before continuing to checkout.
        </Alert>
      ) : null}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <section
          aria-label="Cart items"
          className="rounded-md border border-slate-200 bg-white px-4"
        >
          {items.map((item) => (
            <CartItemRow
              disabled={mutatingItemId === item.item_id}
              item={item}
              key={item.item_id}
              onQuantityChange={(itemId, quantity) => {
                void changeQuantity(itemId, quantity);
              }}
              onRemove={(itemId) => {
                void removeItem(itemId);
              }}
            />
          ))}
        </section>

        <div className="space-y-4">
          <CouponBox
            cartId={cart?.cart_id}
            initialCouponCode={readCartCouponCode()}
            onCouponAccepted={storeCartCouponCode}
            onCouponRejected={clearCartCouponCode}
          />
          <CartSummary
            cart={cart}
            checkoutDisabled={items.length === 0 || hasOutOfStockItem}
          />
        </div>
      </div>
    </div>
  );
}
