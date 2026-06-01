import { z } from 'zod';

export const profileSchema = z.object({
  avatar_url: z
    .string()
    .trim()
    .url('Enter a valid image URL.')
    .optional()
    .or(z.literal('')),
  full_name: z.string().trim().min(2, 'Full name must be at least 2 characters.'),
  phone: z.string().trim().optional().or(z.literal('')),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;

export const notificationPreferencesSchema = z.object({
  email_enabled: z.boolean(),
  marketing_enabled: z.boolean(),
  push_enabled: z.boolean(),
  sms_enabled: z.boolean(),
});

export type NotificationPreferencesFormValues = z.infer<
  typeof notificationPreferencesSchema
>;
