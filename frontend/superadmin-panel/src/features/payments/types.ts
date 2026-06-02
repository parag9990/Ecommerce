export const PAYMENT_STATUSES = [
  "initiated",
  "requires_action",
  "authorized",
  "captured",
  "failed",
  "retry_allowed",
  "partially_refunded",
  "refunded"
] as const;

export type PaymentStatus = (typeof PAYMENT_STATUSES)[number];

export const REFUND_STATUSES = [
  "requested",
  "pending_review",
  "approved",
  "rejected",
  "processing",
  "succeeded",
  "failed"
] as const;

export type RefundStatus = (typeof REFUND_STATUSES)[number];

export const RECONCILIATION_STATUSES = [
  "matched",
  "mismatch",
  "missing_local",
  "missing_provider"
] as const;

export type ReconciliationStatus = (typeof RECONCILIATION_STATUSES)[number];

export const PAYMENT_PROVIDERS = ["stripe", "razorpay", "cod"] as const;

export type PaymentProvider = (typeof PAYMENT_PROVIDERS)[number] | (string & {});

export type Money = {
  amount: number;
  currency: string;
};

export type PaymentAttempt = {
  attempt_id: string;
  payment_id: string;
  status: PaymentStatus;
  provider_reference?: string | null;
  error_code?: string | null;
  error_message?: string | null;
  created_at?: string | null;
};

export type Refund = {
  refund_id: string;
  payment_id: string;
  order_id?: string | null;
  status: RefundStatus;
  amount: Money;
  reason: string;
  requested_by: "buyer" | "seller" | "support" | "admin" | string;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  created_at?: string | null;
};

export type AdminPayment = {
  payment_id: string;
  order_id: string;
  provider: PaymentProvider;
  provider_payment_id?: string | null;
  status: PaymentStatus;
  amount: Money;
  captured_at?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
  attempts?: PaymentAttempt[];
  refunds?: Refund[];
};

export type AdminPaymentDetailResponse = {
  payment: AdminPayment;
  attempts: PaymentAttempt[];
  refunds: Refund[];
};

export type ReconciliationAlert = {
  reconciliation_id: string;
  payment_id?: string | null;
  provider: PaymentProvider;
  status: ReconciliationStatus;
  local_amount?: Money | null;
  provider_amount?: Money | null;
  local_status?: PaymentStatus | string | null;
  provider_status?: string | null;
  settlement_id?: string | null;
  detected_at?: string | null;
  note?: string | null;
};

export type ReconciliationDetailResponse = {
  alert: ReconciliationAlert;
};

export type PaymentFilters = {
  order_id?: string;
  status?: PaymentStatus | "all";
  provider?: PaymentProvider | "all";
  page: number;
  page_size: number;
};

export type RefundFilters = {
  status?: RefundStatus | "all";
  page: number;
  page_size: number;
};

export type ReconciliationFilters = {
  status?: ReconciliationStatus | "all";
  page: number;
  page_size: number;
};

export type AdminPaymentListResponse = {
  payments: AdminPayment[];
  total?: number;
  page?: number;
  page_size?: number;
};

export type RefundListResponse = {
  refunds: Refund[];
  total?: number;
  page?: number;
  page_size?: number;
};

export type ReconciliationListResponse = {
  alerts: ReconciliationAlert[];
  total?: number;
  page?: number;
  page_size?: number;
};

export type RefundReviewDecision = {
  decision: "approved" | "rejected";
  reason: string;
};
