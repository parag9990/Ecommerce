import { apiPost } from '../../../lib/http';
import type { PaymentIntentResponse } from '../types';

type RetryPaymentRequest = {
  idempotency_key: string;
  payment_id?: string | undefined;
};

export function retryPayment(paymentId: string, body: RetryPaymentRequest) {
  return apiPost<PaymentIntentResponse, RetryPaymentRequest>(
    `/api/v1/payments/${encodeURIComponent(paymentId)}/retry`,
    body,
    { auth: true },
  );
}
