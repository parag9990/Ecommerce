import { z } from 'zod';

export const addressSchema = z.object({
  city: z.string().trim().min(2, 'City is required.'),
  country: z.string().trim().min(2, 'Country is required.'),
  is_default: z.boolean(),
  line1: z.string().trim().min(5, 'Address line 1 is required.'),
  line2: z.string().trim().optional().or(z.literal('')),
  name: z.string().trim().min(2, 'Name is required.'),
  phone: z.string().trim().min(8, 'Enter a valid phone number.').optional().or(z.literal('')),
  postal_code: z.string().trim().min(4, 'Postal code is required.'),
  state: z.string().trim().min(2, 'State is required.'),
});

export type AddressFormValues = z.infer<typeof addressSchema>;
