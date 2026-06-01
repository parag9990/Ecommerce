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

export type AddressInput = {
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

export type SuccessResponse = {
  success?: boolean | undefined;
};
