import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import { Button } from '../../../components/ui/button';
import {
  notificationPreferencesSchema,
  type NotificationPreferencesFormValues,
} from '../profile-schema';
import type { NotificationPreference } from '../types';

type NotificationPreferencesFormProps = {
  isSaving?: boolean | undefined;
  onSubmit: (values: NotificationPreferencesFormValues) => Promise<void>;
  preferences: NotificationPreference;
};

const preferenceOptions = [
  {
    description: 'Order, payment, and account messages by email.',
    label: 'Email updates',
    name: 'email_enabled',
  },
  {
    description: 'Important delivery and account alerts by SMS.',
    label: 'SMS alerts',
    name: 'sms_enabled',
  },
  {
    description: 'Browser or app push notifications when available.',
    label: 'Push notifications',
    name: 'push_enabled',
  },
  {
    description: 'Offers, price drops, and promotional updates.',
    label: 'Marketing messages',
    name: 'marketing_enabled',
  },
] as const;

function getDefaultValues(
  preferences: NotificationPreference,
): NotificationPreferencesFormValues {
  return {
    email_enabled: preferences.email_enabled ?? true,
    marketing_enabled: preferences.marketing_enabled ?? false,
    push_enabled: preferences.push_enabled ?? false,
    sms_enabled: preferences.sms_enabled ?? false,
  };
}

export function NotificationPreferencesForm({
  isSaving = false,
  onSubmit,
  preferences,
}: NotificationPreferencesFormProps) {
  const {
    formState: { isDirty },
    handleSubmit,
    register,
    reset,
  } = useForm<NotificationPreferencesFormValues>({
    defaultValues: getDefaultValues(preferences),
    resolver: zodResolver(notificationPreferencesSchema),
  });

  useEffect(() => {
    reset(getDefaultValues(preferences));
  }, [preferences, reset]);

  return (
    <form
      className="space-y-5"
      onSubmit={(event) => {
        void handleSubmit(onSubmit)(event);
      }}
    >
      <div className="divide-y divide-slate-100 rounded-md border border-slate-200 bg-white">
        {preferenceOptions.map((option) => (
          <label
            className="flex items-center justify-between gap-4 px-4 py-4"
            key={option.name}
          >
            <span>
              <span className="block text-sm font-semibold text-slate-950">
                {option.label}
              </span>
              <span className="mt-1 block text-sm leading-6 text-slate-600">
                {option.description}
              </span>
            </span>
            <input
              className="h-5 w-5 rounded border-slate-300 text-blue-600 focus:ring-blue-600"
              type="checkbox"
              {...register(option.name)}
            />
          </label>
        ))}
      </div>

      <div className="flex flex-wrap gap-3">
        <Button disabled={isSaving} type="submit">
          {isSaving ? 'Saving' : 'Save preferences'}
        </Button>
        <Button
          disabled={!isDirty || isSaving}
          onClick={() => {
            reset(getDefaultValues(preferences));
          }}
          type="button"
          variant="secondary"
        >
          Cancel
        </Button>
      </div>
    </form>
  );
}
