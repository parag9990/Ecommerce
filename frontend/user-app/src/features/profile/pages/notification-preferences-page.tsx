import { useEffect, useState } from 'react';

import { Alert } from '../../../components/ui/alert';
import { Button } from '../../../components/ui/button';
import {
  getNotificationPreferences,
  updateNotificationPreferences,
} from '../api/notification-preferences.api';
import { NotificationPreferencesForm } from '../components/notification-preferences-form';
import type { NotificationPreferencesFormValues } from '../profile-schema';
import type { NotificationPreference } from '../types';

type PreferencesState = {
  error?: string | undefined;
  isLoading: boolean;
  preferences?: NotificationPreference | undefined;
  success?: string | undefined;
};

function getErrorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function PreferencesSkeleton() {
  return (
    <div className="space-y-3 rounded-md border border-slate-200 bg-white p-5">
      {Array.from({ length: 4 }, (_, index) => (
        <div className="flex items-center justify-between gap-4" key={index}>
          <div className="space-y-2">
            <div className="h-4 w-40 animate-pulse rounded bg-slate-200" />
            <div className="h-3 w-64 animate-pulse rounded bg-slate-200" />
          </div>
          <div className="h-5 w-5 animate-pulse rounded bg-slate-200" />
        </div>
      ))}
    </div>
  );
}

export function NotificationPreferencesPage() {
  const [state, setState] = useState<PreferencesState>({ isLoading: true });
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    const controller = new AbortController();

    async function loadPreferences() {
      setState((current) => ({
        ...current,
        error: undefined,
        isLoading: true,
      }));

      try {
        const preferences = await getNotificationPreferences(controller.signal);

        if (!controller.signal.aborted) {
          setState({ isLoading: false, preferences });
        }
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return;
        }

        setState({
          error: getErrorMessage(
            error,
            'Notification preferences could not be loaded.',
          ),
          isLoading: false,
        });
      }
    }

    void loadPreferences();

    return () => {
      controller.abort();
    };
  }, []);

  async function savePreferences(values: NotificationPreferencesFormValues) {
    setIsSaving(true);
    setState((current) => ({ ...current, error: undefined, success: undefined }));

    try {
      const preferences = await updateNotificationPreferences(values);

      setState({
        isLoading: false,
        preferences,
        success: 'Notification preferences saved.',
      });
    } catch (error) {
      setState((current) => ({
        ...current,
        error: getErrorMessage(error, 'Preferences could not be saved.'),
      }));
    } finally {
      setIsSaving(false);
    }
  }

  if (state.isLoading) {
    return (
      <section className="space-y-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-950">
            Notifications
          </h1>
          <p className="mt-1 text-sm text-slate-600">
            Choose how account and shopping updates reach you.
          </p>
        </div>
        <PreferencesSkeleton />
      </section>
    );
  }

  if (!state.preferences) {
    return (
      <div className="space-y-4">
        <Alert title="Preferences could not be loaded" variant="error">
          {state.error ?? 'Please try again.'}
        </Alert>
        <Button
          onClick={() => {
            window.location.reload();
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
          Notifications
        </h1>
        <p className="mt-1 text-sm text-slate-600">
          Transactional updates stay separate from marketing messages.
        </p>
      </div>

      {state.error ? <Alert variant="error">{state.error}</Alert> : null}
      {state.success ? (
        <Alert variant="success">{state.success}</Alert>
      ) : null}

      <NotificationPreferencesForm
        isSaving={isSaving}
        onSubmit={savePreferences}
        preferences={state.preferences}
      />
    </section>
  );
}
