import { z } from 'zod';

export const cancelOrderSchema = z.object({
  reason: z.string().trim().min(5, 'Enter at least 5 characters.'),
});

export type CancelOrderFormValues = z.infer<typeof cancelOrderSchema>;
