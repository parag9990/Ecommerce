import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Input } from '../../../components/ui/input';
import { PasswordInput } from '../../../components/ui/password-input';
import { routePaths } from '../../../routes/route-paths';
import { forgotPassword, resetPassword } from '../api/auth.api';
import { AuthCard } from '../components/auth-card';
import { AuthField } from '../components/auth-field';
import { AuthPageFrame } from '../components/auth-page-frame';
import { AuthSubmit } from '../components/auth-submit';
import { PasswordStrength } from '../components/password-strength';
import { ResendOtpButton } from '../components/resend-otp-button';
import {
  resetPasswordSchema,
  type ResetPasswordFormValues,
} from '../schemas';
import { toTrimmedValue } from '../utils';

const fieldIds = {
  confirmPassword: 'reset-confirm-password',
  newPassword: 'reset-new-password',
  otp: 'reset-otp',
} as const;

export function ResetPasswordPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialChallengeId = searchParams.get('challenge_id') ?? '';
  const target = searchParams.get('target') ?? '';
  const [serverError, setServerError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [challengeId, setChallengeId] = useState(initialChallengeId);
  const [newPassword, setNewPassword] = useState('');

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    setValue,
  } = useForm<ResetPasswordFormValues>({
    defaultValues: {
      challenge_id: initialChallengeId,
      confirmPassword: '',
      new_password: '',
      otp: '',
    },
    resolver: zodResolver(resetPasswordSchema),
  });

  const hasOtpError = Boolean(errors.otp ?? errors.challenge_id);
  const newPasswordRegistration = register('new_password');

  async function onSubmit(values: ResetPasswordFormValues) {
    setServerError(null);
    setSuccessMessage(null);

    try {
      await resetPassword({
        challenge_id: toTrimmedValue(values.challenge_id),
        new_password: values.new_password,
        otp: toTrimmedValue(values.otp),
      });

      await navigate(`${routePaths.login}?passwordReset=1`, { replace: true });
    } catch (error) {
      setServerError(
        error instanceof Error ? error.message : 'Password reset failed.',
      );
    }
  }

  async function handleResend() {
    setServerError(null);
    setSuccessMessage(null);

    if (!target) {
      setServerError('Email or phone is required before resending OTP.');
      throw new Error('Missing password reset target.');
    }

    try {
      const challenge = await forgotPassword({ identifier: target });

      setValue('challenge_id', challenge.challenge_id, {
        shouldDirty: true,
        shouldValidate: true,
      });
      setChallengeId(challenge.challenge_id);
      setSuccessMessage(`New reset OTP sent. It expires in ${challenge.expires_in}s.`);
    } catch (error) {
      setServerError(
        error instanceof Error ? error.message : 'Could not resend reset OTP.',
      );
      throw error;
    }
  }

  return (
    <AuthPageFrame>
      <AuthCard
        description="Enter the OTP and choose a new password."
        title="Reset password"
      >
        <form
          className="space-y-4"
          noValidate
          onSubmit={(event) => {
            void handleSubmit(onSubmit)(event);
          }}
        >
          {!challengeId ? (
            <Alert variant="info">
              Password reset challenge is missing. Request a new reset OTP.
            </Alert>
          ) : null}

          {serverError ? <Alert variant="error">{serverError}</Alert> : null}

          {successMessage ? (
            <Alert variant="success">{successMessage}</Alert>
          ) : null}

          <input type="hidden" {...register('challenge_id')} />

          <AuthField
            error={errors.otp?.message ?? errors.challenge_id?.message}
            errorId={`${fieldIds.otp}-error`}
            htmlFor={fieldIds.otp}
            label="OTP"
          >
            <Input
              aria-describedby={hasOtpError ? `${fieldIds.otp}-error` : undefined}
              aria-invalid={hasOtpError}
              autoComplete="one-time-code"
              id={fieldIds.otp}
              inputMode="numeric"
              invalid={hasOtpError}
              maxLength={6}
              pattern="[0-9]*"
              type="text"
              {...register('otp')}
            />
          </AuthField>

          <AuthField
            error={errors.new_password?.message}
            errorId={`${fieldIds.newPassword}-error`}
            htmlFor={fieldIds.newPassword}
            label="New password"
          >
            <PasswordInput
              aria-describedby={
                errors.new_password
                  ? `${fieldIds.newPassword}-error`
                  : undefined
              }
              aria-invalid={Boolean(errors.new_password)}
              autoComplete="new-password"
              id={fieldIds.newPassword}
              invalid={Boolean(errors.new_password)}
              {...newPasswordRegistration}
              onChange={(event) => {
                setNewPassword(event.target.value);
                void newPasswordRegistration.onChange(event);
              }}
            />
          </AuthField>

          <PasswordStrength password={newPassword} />

          <AuthField
            error={errors.confirmPassword?.message}
            errorId={`${fieldIds.confirmPassword}-error`}
            htmlFor={fieldIds.confirmPassword}
            label="Confirm new password"
          >
            <PasswordInput
              aria-describedby={
                errors.confirmPassword
                  ? `${fieldIds.confirmPassword}-error`
                  : undefined
              }
              aria-invalid={Boolean(errors.confirmPassword)}
              autoComplete="new-password"
              id={fieldIds.confirmPassword}
              invalid={Boolean(errors.confirmPassword)}
              {...register('confirmPassword')}
            />
          </AuthField>

          <AuthSubmit isSubmitting={isSubmitting}>Reset password</AuthSubmit>

          <ResendOtpButton disabled={!target} onResend={handleResend} />
        </form>

        <p className="mt-6 text-center text-sm text-slate-600">
          Back to{' '}
          <Link
            className="font-medium text-blue-700 hover:text-blue-800"
            to={routePaths.login}
          >
            login
          </Link>
        </p>
      </AuthCard>
    </AuthPageFrame>
  );
}
