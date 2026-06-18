import { z } from 'zod';

export const checkoutSchema = z.object({
  address_id: z.string().trim().min(1, 'Select a delivery address.'),
  coupon_code: z
    .string()
    .trim()
    .max(40, 'Coupon code must be 40 characters or fewer.')
    .optional(),
  payment_provider: z.string().trim().min(1, 'Select a payment method.'),
});

export type CheckoutFormValues = z.infer<typeof checkoutSchema>;
