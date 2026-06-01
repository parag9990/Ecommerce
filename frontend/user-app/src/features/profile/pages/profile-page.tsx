import { useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import { ProfileForm } from '../components/profile-form';
import { useProfileMutation } from '../hooks/use-profile-mutation';
import { useProfileQuery } from '../hooks/use-profile-query';
import type { ProfileFormValues } from '../profile-schema';

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function cleanOptional(value?: string) {
  const trimmed = value?.trim();

  return trimmed ? trimmed : undefined;
}

function ProfileSkeleton() {
  return (
    <div className="space-y-5 rounded-md border border-slate-200 bg-white p-5">
      {Array.from({ length: 4 }, (_, index) => (
        <div className="space-y-2" key={index}>
          <div className="h-4 w-28 animate-pulse rounded bg-slate-200" />
          <div className="h-11 animate-pulse rounded bg-slate-200" />
        </div>
      ))}
    </div>
  );
}

export function ProfilePage() {
  const profileQuery = useProfileQuery();
  const profileMutation = useProfileMutation();
  const [actionError, setActionError] = useState<string>();
  const [success, setSuccess] = useState<string>();

  async function saveProfile(values: ProfileFormValues) {
    setActionError(undefined);
    setSuccess(undefined);

    try {
      await profileMutation.mutateAsync({
        avatar_url: cleanOptional(values.avatar_url),
        full_name: values.full_name.trim(),
        phone: cleanOptional(values.phone),
      });
      setSuccess('Profile saved.');
    } catch (error) {
      setActionError(getErrorMessage(error, 'Profile could not be saved.'));
    }
  }

  if (profileQuery.isLoading) {
    return (
      <section className="space-y-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
            Profile
          </h1>
          <p className="mt-1 text-sm text-slate-600">
            Manage your buyer account details.
          </p>
        </div>
        <ProfileSkeleton />
      </section>
    );
  }

  const loadError =
    profileQuery.error instanceof Error
      ? profileQuery.error.message
      : profileQuery.isError
        ? 'Profile could not be loaded.'
        : undefined;

  if (!profileQuery.data) {
    return (
      <div className="space-y-4">
        <Alert title="Profile could not be loaded" variant="error">
          {loadError ?? 'Please try again.'}
        </Alert>
        <Button
          onClick={() => {
            void profileQuery.refetch();
          }}
          type="button"
          variant="secondary"
        >
          Retry
        </Button>
      </div>
    );
  }

  return (
    <section className="space-y-5">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
          Profile
        </h1>
        <p className="mt-1 text-sm text-slate-600">
          Keep your contact details current for orders and delivery updates.
        </p>
      </div>

      {actionError ? <Alert variant="error">{actionError}</Alert> : null}
      {success ? <Alert variant="success">{success}</Alert> : null}

      <div className="rounded-md border border-slate-200 bg-white p-5">
        <ProfileForm
          isSaving={profileMutation.isPending}
          onSubmit={saveProfile}
          profile={profileQuery.data}
        />
      </div>
    </section>
  );
}
