import { zodResolver } from '@hookform/resolvers/zod';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';

import { Button } from '../../../components/ui/button';
import { Input } from '../../../components/ui/input';
import {
  profileSchema,
  type ProfileFormValues,
} from '../profile-schema';
import type { UserProfile } from '../types';

type ProfileFormProps = {
  isSaving?: boolean | undefined;
  onSubmit: (values: ProfileFormValues) => Promise<void>;
  profile: UserProfile;
};

function getDefaultValues(profile: UserProfile): ProfileFormValues {
  return {
    avatar_url: profile.avatar_url ?? '',
    full_name: profile.full_name ?? '',
    phone: profile.phone ?? '',
  };
}

export function ProfileForm({
  isSaving = false,
  onSubmit,
  profile,
}: ProfileFormProps) {
  const {
    formState: { errors, isDirty },
    handleSubmit,
    register,
    reset,
  } = useForm<ProfileFormValues>({
    defaultValues: getDefaultValues(profile),
    resolver: zodResolver(profileSchema),
  });

  useEffect(() => {
    reset(getDefaultValues(profile));
  }, [profile, reset]);

  return (
    <form
      className="space-y-5"
      noValidate
      onSubmit={(event) => {
        void handleSubmit(onSubmit)(event);
      }}
    >
      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Email</span>
        <Input
          readOnly
          value={profile.email ?? ''}
          className="bg-slate-100 text-slate-600"
        />
      </label>

      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Full name</span>
        <Input
          autoComplete="name"
          invalid={Boolean(errors.full_name)}
          {...register('full_name')}
        />
        {errors.full_name ? (
          <span className="text-sm text-red-700">{errors.full_name.message}</span>
        ) : null}
      </label>

      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Phone</span>
        <Input
          autoComplete="tel"
          invalid={Boolean(errors.phone)}
          placeholder="+1 555 123 4567"
          {...register('phone')}
        />
        {errors.phone ? (
          <span className="text-sm text-red-700">{errors.phone.message}</span>
        ) : null}
      </label>

      <label className="grid gap-1.5">
        <span className="text-sm font-medium text-slate-700">Avatar URL</span>
        <Input
          autoComplete="url"
          invalid={Boolean(errors.avatar_url)}
          placeholder="https://example.com/avatar.jpg"
          {...register('avatar_url')}
        />
        {errors.avatar_url ? (
          <span className="text-sm text-red-700">
            {errors.avatar_url.message}
          </span>
        ) : null}
      </label>

      <div className="flex flex-wrap gap-3">
        <Button disabled={isSaving} type="submit">
          {isSaving ? 'Saving' : 'Save profile'}
        </Button>
        <Button
          disabled={!isDirty || isSaving}
          onClick={() => {
            reset(getDefaultValues(profile));
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
