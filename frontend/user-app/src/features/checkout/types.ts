import type { Cart, Money } from '../cart/types';

export type Address = {
  address_id?: string | undefined;
  city: string;
  country: string;
  is_default?: boolean | undefined;
  line1: string;
  line2?: string | undefined;
  name: string;
  phone?: string | undefined;
  postal_code: string;
  state: string;
};

export type AddressListResponse = {
  addresses?: Address[] | undefined;
};

export type CheckoutRequest = {
  address_id: string;
  coupon_code?: string | undefined;
  idempotency_key: string;
  payment_provider: string;
};

export type Order = {
  created_at?: string | undefined;
  items?: unknown[] | undefined;
  order_id?: string | undefined;
  status?: string | undefined;
  total?: Money | undefined;
  user_id?: string | undefined;
};

export type PaymentIntentResponse = {
  amount?: Money | undefined;
  client_secret?: string | undefined;
  next_action_url?: string | undefined;
  payment_id?: string | undefined;
  payment_url?: string | undefined;
  provider?: string | undefined;
  redirect_url?: string | undefined;
  status?: string | undefined;
};

export type CheckoutResponse = {
  order?: Order | undefined;
  payment_intent?: PaymentIntentResponse | undefined;
};

export type CheckoutPageData = {
  addresses?: Address[] | undefined;
  cart?: Cart | undefined;
};
