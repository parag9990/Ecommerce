import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { Link } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { EmptyState } from '../../../components/ui/empty-state';
import { env } from '../../../lib/env';
import { queryKeys } from '../../../lib/query-keys';
import { routePaths } from '../../../routes/route-paths';
import { readCartCouponCode } from '../../cart/cart-coupon';
import { useCartQuery } from '../../cart/hooks/use-cart-query';
import { useAddressesQuery } from '../../addresses/hooks/use-addresses-query';
import { createCheckout } from '../api/checkout.api';
import { checkoutSchema, type CheckoutFormValues } from '../checkout-schema';
import { getCheckoutIdempotencyKey } from '../checkout-idempotency';
import { AddressStep } from '../components/address-step';
import { CheckoutProgress } from '../components/checkout-progress';
import { OrderReviewStep } from '../components/order-review-step';
import { PaymentStep } from '../components/payment-step';
import { ProviderPaymentWidget } from '../components/provider-payment-widget';
import type { PaymentIntentResponse } from '../types';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function CheckoutPage() {
  const [error, setError] = useState<string>();
  const [orderId, setOrderId] = useState<string>();
  const [paymentIntent, setPaymentIntent] = useState<PaymentIntentResponse>();
  const queryClient = useQueryClient();
  const cartQuery = useCartQuery();
  const addressesQuery = useAddressesQuery();
  const checkoutMutation = useMutation({
    mutationFn: createCheckout,
    onSuccess: (response) => {
      if (response.order?.order_id) {
        void queryClient.invalidateQueries({ queryKey: queryKeys.orders.all });
      }
    },
  });
  const defaultPaymentProvider = env.paymentProviders[0] ?? 'stripe';
  const {
    control,
    formState: { errors, isSubmitting },
    handleSubmit,
    setValue,
  } = useForm<CheckoutFormValues>({
    defaultValues: {
      address_id: '',
      coupon_code: readCartCouponCode(),
      payment_provider: defaultPaymentProvider,
    },
    resolver: zodResolver(checkoutSchema),
  });
  const selectedAddressId = useWatch({ control, name: 'address_id' });
  const selectedPaymentProvider = useWatch({
    control,
    name: 'payment_provider',
  });
  const couponCode = useWatch({ control, name: 'coupon_code' });
  const addresses = useMemo(
    () => addressesQuery.data ?? [],
    [addressesQuery.data],
  );
  const cart = cartQuery.data;
  const items = cart?.items ?? [];
  const hasOutOfStockItem = items.some(
    (item) => item.stock_status === 'out_of_stock',
  );
  const submitDisabled =
    isSubmitting ||
    checkoutMutation.isPending ||
    addresses.length === 0 ||
    items.length === 0 ||
    hasOutOfStockItem;

  useEffect(() => {
    if (selectedAddressId || addresses.length === 0) {
      return;
    }

    const defaultAddress =
      addresses.find((address) => address.is_default) ?? addresses[0];

    if (defaultAddress?.address_id) {
      setValue('address_id', defaultAddress.address_id, {
        shouldValidate: true,
      });
    }
  }, [addresses, selectedAddressId, setValue]);

  async function submitCheckout(values: CheckoutFormValues) {
    setError(undefined);

    try {
      const response = await checkoutMutation.mutateAsync({
        address_id: values.address_id,
        coupon_code: values.coupon_code?.trim().toUpperCase() || undefined,
        idempotency_key: getCheckoutIdempotencyKey(),
        payment_provider: values.payment_provider,
      });

      if (!response.payment_intent) {
        throw new Error('Payment intent was not returned by checkout.');
      }

      setOrderId(response.order?.order_id);
      setPaymentIntent(response.payment_intent);
    } catch (submitError) {
      setError(getErrorMessage(submitError, 'Checkout could not be created.'));
    }
  }

  const loadError =
    cartQuery.error instanceof Error
      ? cartQuery.error.message
      : addressesQuery.error instanceof Error
        ? addressesQuery.error.message
        : cartQuery.isError || addressesQuery.isError
          ? 'Checkout could not be loaded.'
          : undefined;
  const pageError = error ?? loadError;
  const isLoading = cartQuery.isLoading || addressesQuery.isLoading;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <CheckoutProgress activeStep="address" />
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
          <div className="space-y-4">
            <div className="h-36 animate-pulse rounded-md bg-slate-200" />
            <div className="h-36 animate-pulse rounded-md bg-slate-200" />
          </div>
          <div className="h-64 animate-pulse rounded-md bg-slate-200" />
        </div>
      </div>
    );
  }

  if (paymentIntent) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <CheckoutProgress activeStep="payment" />
        <ProviderPaymentWidget paymentIntent={paymentIntent} />
        {orderId ? (
          <p className="text-center text-sm text-slate-500">
            Order reference: {orderId}
          </p>
        ) : null}
      </div>
    );
  }

  if (items.length === 0 && !pageError) {
    return (
      <EmptyState
        action={
          <Link
            className="inline-flex h-10 items-center justify-center rounded-md bg-blue-600 px-4 text-sm font-semibold text-white transition hover:bg-blue-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
            to={routePaths.home}
          >
            Continue shopping
          </Link>
        }
        description="Add products before starting checkout."
        title="Your cart is empty"
      />
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Checkout
        </h1>
        <CheckoutProgress activeStep="review" />
      </div>

      {pageError ? <Alert variant="error">{pageError}</Alert> : null}
      {hasOutOfStockItem ? (
        <Alert variant="error">
          Remove out-of-stock items from your cart before checkout.
        </Alert>
      ) : null}

      <form
        className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]"
        noValidate
        onSubmit={(event) => {
          void handleSubmit(submitCheckout)(event);
        }}
      >
        <div className="space-y-6">
          <AddressStep
            addresses={addresses}
            onSelect={(addressId) => {
              setValue('address_id', addressId, { shouldValidate: true });
            }}
            selectedAddressId={selectedAddressId}
          />
          {errors.address_id ? (
            <p className="text-sm text-red-700">{errors.address_id.message}</p>
          ) : null}

          <PaymentStep
            onProviderChange={(provider) => {
              setValue('payment_provider', provider, { shouldValidate: true });
            }}
            selectedProvider={selectedPaymentProvider}
          />
          {errors.payment_provider ? (
            <p className="text-sm text-red-700">
              {errors.payment_provider.message}
            </p>
          ) : null}
        </div>

        <aside className="space-y-4 rounded-md border border-slate-200 bg-white p-4">
          <OrderReviewStep cart={cart} couponCode={couponCode} />
          <Button disabled={submitDisabled} fullWidth type="submit">
            {isSubmitting || checkoutMutation.isPending
              ? 'Creating payment'
              : 'Create payment'}
          </Button>
        </aside>
      </form>
    </div>
  );
}
