import { apiFetch } from "../../../lib/http";
import type {
  AdminPayment,
  AdminPaymentDetailResponse,
  AdminPaymentListResponse,
  Money,
  PaymentAttempt,
  PaymentFilters,
  PaymentStatus,
  ReconciliationAlert,
  ReconciliationDetailResponse,
  ReconciliationFilters,
  ReconciliationListResponse,
  ReconciliationStatus,
  Refund,
  RefundFilters,
  RefundListResponse,
  RefundReviewDecision,
  RefundStatus
} from "../types";
import { PAYMENT_STATUSES, RECONCILIATION_STATUSES, REFUND_STATUSES } from "../types";

const PAYMENTS_PATH = "/api/v1/admin/payments";
const REFUNDS_PATH = "/api/v1/admin/refunds";
const RECONCILIATIONS_PATH = "/api/v1/admin/payment-reconciliations";
const DEFAULT_CURRENCY = "INR";
const DEFAULT_PAYMENT_STATUS: PaymentStatus = "initiated";
const DEFAULT_REFUND_STATUS: RefundStatus = "requested";
const DEFAULT_RECONCILIATION_STATUS: ReconciliationStatus = "mismatch";

type RawPayment = Partial<Omit<AdminPayment, "amount" | "status" | "attempts" | "refunds">> & {
  amount?: Partial<Money> | null;
  status?: string | null;
  attempts?: Array<Partial<PaymentAttempt>>;
  refunds?: Array<Partial<Refund>>;
};

type RawPaymentDetailResponse = Partial<Omit<AdminPaymentDetailResponse, "payment" | "attempts" | "refunds">> & {
  payment?: RawPayment | null;
  attempts?: Array<Partial<PaymentAttempt>>;
  refunds?: Array<Partial<Refund>>;
};

type RawRefund = Partial<Omit<Refund, "amount" | "status">> & {
  amount?: Partial<Money> | null;
  status?: string | null;
};

type RawReconciliationAlert = Partial<
  Omit<ReconciliationAlert, "local_amount" | "provider_amount" | "status">
> & {
  local_amount?: Partial<Money> | null;
  provider_amount?: Partial<Money> | null;
  status?: string | null;
};

type RawReconciliationDetailResponse = Partial<Omit<ReconciliationDetailResponse, "alert">> & {
  alert?: RawReconciliationAlert | null;
};

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function appendQuery(path: string, query: string): string {
  return query ? `${path}?${query}` : path;
}

function isPaymentStatus(status?: string | null): status is PaymentStatus {
  return PAYMENT_STATUSES.includes(status as PaymentStatus);
}

function isRefundStatus(status?: string | null): status is RefundStatus {
  return REFUND_STATUSES.includes(status as RefundStatus);
}

function isReconciliationStatus(status?: string | null): status is ReconciliationStatus {
  return RECONCILIATION_STATUSES.includes(status as ReconciliationStatus);
}

function normalizeMoney(money?: Partial<Money> | null): Money {
  return {
    amount: typeof money?.amount === "number" ? money.amount : 0,
    currency: money?.currency?.trim() || DEFAULT_CURRENCY
  };
}

function normalizeAttempt(attempt: Partial<PaymentAttempt>, index: number, paymentId: string): PaymentAttempt {
  return {
    attempt_id: attempt.attempt_id?.trim() || `attempt_${index + 1}`,
    payment_id: attempt.payment_id?.trim() || paymentId,
    status: isPaymentStatus(attempt.status) ? attempt.status : DEFAULT_PAYMENT_STATUS,
    provider_reference: attempt.provider_reference?.trim() || null,
    error_code: attempt.error_code?.trim() || null,
    error_message: attempt.error_message?.trim() || null,
    created_at: attempt.created_at ?? null
  };
}

function normalizeRefund(refund: RawRefund, index = 0, paymentId?: string): Refund {
  return {
    refund_id: refund.refund_id?.trim() || `refund_${index + 1}`,
    payment_id: refund.payment_id?.trim() || paymentId || "unknown_payment",
    order_id: refund.order_id?.trim() || null,
    status: isRefundStatus(refund.status) ? refund.status : DEFAULT_REFUND_STATUS,
    amount: normalizeMoney(refund.amount),
    reason: refund.reason?.trim() || "No reason provided.",
    requested_by: refund.requested_by?.trim() || "support",
    reviewed_by: refund.reviewed_by?.trim() || null,
    reviewed_at: refund.reviewed_at ?? null,
    created_at: refund.created_at ?? null
  };
}

function normalizePayment(payment: RawPayment): AdminPayment {
  const paymentId = payment.payment_id?.trim() || "unknown_payment";

  return {
    ...payment,
    payment_id: paymentId,
    order_id: payment.order_id?.trim() || "unknown_order",
    provider: payment.provider?.trim() || "unknown",
    provider_payment_id: payment.provider_payment_id?.trim() || null,
    status: isPaymentStatus(payment.status) ? payment.status : DEFAULT_PAYMENT_STATUS,
    amount: normalizeMoney(payment.amount),
    captured_at: payment.captured_at ?? null,
    created_at: payment.created_at ?? null,
    updated_at: payment.updated_at ?? null,
    attempts: (payment.attempts ?? []).map((attempt, index) => normalizeAttempt(attempt, index, paymentId)),
    refunds: (payment.refunds ?? []).map((refund, index) => normalizeRefund(refund, index, paymentId))
  };
}

function normalizeReconciliationAlert(
  alert: RawReconciliationAlert,
  index = 0
): ReconciliationAlert {
  return {
    reconciliation_id: alert.reconciliation_id?.trim() || `reconciliation_${index + 1}`,
    payment_id: alert.payment_id?.trim() || null,
    provider: alert.provider?.trim() || "unknown",
    status: isReconciliationStatus(alert.status) ? alert.status : DEFAULT_RECONCILIATION_STATUS,
    local_amount: alert.local_amount ? normalizeMoney(alert.local_amount) : null,
    provider_amount: alert.provider_amount ? normalizeMoney(alert.provider_amount) : null,
    local_status: alert.local_status?.trim() || null,
    provider_status: alert.provider_status?.trim() || null,
    settlement_id: alert.settlement_id?.trim() || null,
    detected_at: alert.detected_at ?? null,
    note: alert.note?.trim() || null
  };
}

export function toAdminPaymentsQueryString(filters: PaymentFilters): string {
  return buildQueryString({
    page: filters.page,
    page_size: filters.page_size,
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    provider: filters.provider && filters.provider !== "all" ? filters.provider : undefined,
    order_id: filters.order_id?.trim()
  });
}

export function toRefundsQueryString(filters: RefundFilters): string {
  return buildQueryString({
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    page: filters.page,
    page_size: filters.page_size
  });
}

export function toReconciliationQueryString(filters: ReconciliationFilters): string {
  return buildQueryString({
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    page: filters.page,
    page_size: filters.page_size
  });
}

export async function listAdminPayments(filters: PaymentFilters): Promise<AdminPaymentListResponse> {
  const query = toAdminPaymentsQueryString(filters);
  const response = await apiFetch<AdminPaymentListResponse>(appendQuery(PAYMENTS_PATH, query));

  return {
    ...response,
    payments: (response.payments ?? []).map((payment) => normalizePayment(payment)),
    page: response.page ?? filters.page,
    page_size: response.page_size ?? filters.page_size
  };
}

export async function getAdminPayment(paymentId: string): Promise<AdminPaymentDetailResponse> {
  const response = await apiFetch<RawPaymentDetailResponse>(
    `${PAYMENTS_PATH}/${encodeURIComponent(paymentId)}`
  );
  const payment = normalizePayment(response.payment ?? { payment_id: paymentId });

  return {
    ...response,
    payment,
    attempts: (response.attempts ?? payment.attempts ?? []).map((attempt, index) =>
      normalizeAttempt(attempt, index, payment.payment_id)
    ),
    refunds: (response.refunds ?? payment.refunds ?? []).map((refund, index) =>
      normalizeRefund(refund, index, payment.payment_id)
    )
  };
}

export async function listRefunds(filters: RefundFilters): Promise<RefundListResponse> {
  const query = toRefundsQueryString(filters);
  const response = await apiFetch<RefundListResponse>(appendQuery(REFUNDS_PATH, query));

  return {
    ...response,
    refunds: (response.refunds ?? []).map((refund, index) => normalizeRefund(refund, index)),
    page: response.page ?? filters.page,
    page_size: response.page_size ?? filters.page_size
  };
}

export async function reviewRefund(refundId: string, input: RefundReviewDecision): Promise<Refund> {
  const response = await apiFetch<RawRefund>(`${REFUNDS_PATH}/${encodeURIComponent(refundId)}/review`, {
    method: "POST",
    body: JSON.stringify({
      decision: input.decision,
      reason: input.reason
    })
  });

  return normalizeRefund(response, 0, response.payment_id);
}

export async function listReconciliationAlerts(
  filters: ReconciliationFilters
): Promise<ReconciliationListResponse> {
  const query = toReconciliationQueryString(filters);
  const response = await apiFetch<ReconciliationListResponse>(appendQuery(RECONCILIATIONS_PATH, query));

  return {
    ...response,
    alerts: (response.alerts ?? []).map(normalizeReconciliationAlert),
    page: response.page ?? filters.page,
    page_size: response.page_size ?? filters.page_size
  };
}

export async function getReconciliationAlert(
  reconciliationId: string
): Promise<ReconciliationDetailResponse> {
  const response = await apiFetch<RawReconciliationDetailResponse>(
    `${RECONCILIATIONS_PATH}/${encodeURIComponent(reconciliationId)}`
  );

  return {
    ...response,
    alert: normalizeReconciliationAlert(response.alert ?? { reconciliation_id: reconciliationId })
  };
}
