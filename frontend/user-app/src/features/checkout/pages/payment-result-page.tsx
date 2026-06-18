import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { queryKeys } from '../../../lib/query-keys';
import { clearCartCouponCode } from '../../cart/cart-coupon';
import { useOrderDetailQuery } from '../../orders/hooks/use-order-detail-query';
import { retryPayment } from '../api/payment.api';
import {
  clearCheckoutIdempotencyKey,
  resetCheckoutIdempotencyKey,
} from '../checkout-idempotency';
import { PaymentStatusCard } from '../components/payment-status-card';
import { ProviderPaymentWidget } from '../components/provider-payment-widget';
import type { PaymentIntentResponse } from '../types';

type PaymentResultPageProps = {
  result: 'success' | 'failure';
};

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function PaymentResultPage({ result }: PaymentResultPageProps) {
  const [params] = useSearchParams();
  const orderId = params.get('order_id') ?? '';
  const paymentId = params.get('payment_id') ?? '';
  const queryClient = useQueryClient();
  const orderQuery = useOrderDetailQuery(orderId);
  const [retryIntent, setRetryIntent] = useState<PaymentIntentResponse>();
  const [retrying, setRetrying] = useState(false);
  const [retryError, setRetryError] = useState<string>();

  useEffect(() => {
    if (!orderId || result !== 'success' || !orderQuery.data) {
      return;
    }

    clearCheckoutIdempotencyKey();
    clearCartCouponCode();
    queryClient.removeQueries({ queryKey: queryKeys.cart.all });
    void queryClient.invalidateQueries({ queryKey: queryKeys.orders.all });
  }, [orderId, orderQuery.data, queryClient, result]);

  async function handleRetry() {
    if (!paymentId) {
      setRetryError('Payment reference is missing.');
      return;
    }

    setRetryError(undefined);
    setRetrying(true);

    try {
      const intent = await retryPayment(paymentId, {
        idempotency_key: resetCheckoutIdempotencyKey(),
        payment_id: paymentId,
      });
      setRetryIntent(intent);
    } catch (retryFailure) {
      setRetryError(
        getErrorMessage(retryFailure, 'Payment retry could not be started.'),
      );
    } finally {
      setRetrying(false);
    }
  }

  if (!orderId) {
    return (
      <div className="mx-auto max-w-2xl">
        <Alert title="Order reference missing" variant="error">
          Open checkout again to confirm the latest order status.
        </Alert>
      </div>
    );
  }

  if (retryIntent) {
    return (
      <div className="mx-auto max-w-3xl space-y-6">
        <ProviderPaymentWidget paymentIntent={retryIntent} />
      </div>
    );
  }

  const error =
    orderQuery.error instanceof Error
      ? orderQuery.error.message
      : orderQuery.isError
        ? 'Order status could not be loaded.'
        : undefined;

  return (
    <div className="space-y-4">
      <PaymentStatusCard
        error={error}
        isLoading={orderQuery.isLoading}
        onRetry={
          result === 'failure'
            ? () => {
                void handleRetry();
              }
            : undefined
        }
        order={orderQuery.data}
        result={result}
        retrying={retrying}
      />
      {retryError ? (
        <div className="mx-auto max-w-2xl">
          <Alert variant="error">{retryError}</Alert>
        </div>
      ) : null}
    </div>
  );
}
