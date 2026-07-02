import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useNavigate, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Input } from '../../../components/ui/input';
import { routePaths } from '../../../routes/route-paths';
import { sendOtp, verifyOtp } from '../api/auth.api';
import { AuthCard } from '../components/auth-card';
import { AuthField } from '../components/auth-field';
import { AuthPageFrame } from '../components/auth-page-frame';
import { AuthSubmit } from '../components/auth-submit';
import { ResendOtpButton } from '../components/resend-otp-button';
import { otpSchema, type OtpFormValues } from '../schemas';
import {
  detectOtpChannel,
  isOtpPurpose,
  localEmailOtpNotice,
  otpDeliveryErrorMessage,
  otpDeliverySuccessMessage,
  toTrimmedValue,
} from '../utils';

const fieldIds = {
  otp: 'otp-code',
} as const;

function readOtpPurpose(searchParams: URLSearchParams) {
  const purpose = searchParams.get('purpose');

  return isOtpPurpose(purpose) ? purpose : 'signup';
}

export function OtpPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const target = searchParams.get('target') ?? '';
  const initialChallengeId = searchParams.get('challenge_id') ?? '';
  const purpose = readOtpPurpose(searchParams);
  const [serverError, setServerError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [challengeId, setChallengeId] = useState(initialChallengeId);
  const deliveryNotice =
    challengeId && !successMessage ? localEmailOtpNotice(target) : null;

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
    setValue,
  } = useForm<OtpFormValues>({
    defaultValues: {
      challenge_id: initialChallengeId,
      otp: '',
    },
    resolver: zodResolver(otpSchema),
  });

  const hasOtpError = Boolean(errors.otp ?? errors.challenge_id);

  async function onSubmit(values: OtpFormValues) {
    setServerError(null);
    setSuccessMessage(null);

    try {
      await verifyOtp({
        challenge_id: toTrimmedValue(values.challenge_id),
        otp: toTrimmedValue(values.otp),
      });

      await navigate(`${routePaths.login}?verified=1`, { replace: true });
    } catch (error) {
      setServerError(
        error instanceof Error ? error.message : 'OTP verification failed.',
      );
    }
  }

  async function handleResend() {
    setServerError(null);
    setSuccessMessage(null);

    if (!target) {
      setServerError('Email or phone is required before resending OTP.');
      throw new Error('Missing OTP target.');
    }

    try {
      const challenge = await sendOtp({
        channel: detectOtpChannel(target),
        purpose,
        target,
      });

      setValue('challenge_id', challenge.challenge_id, {
        shouldDirty: true,
        shouldValidate: true,
      });
      setChallengeId(challenge.challenge_id);
      setSuccessMessage(
        otpDeliverySuccessMessage(target, challenge.expires_in),
      );
    } catch (error) {
      setServerError(
        otpDeliveryErrorMessage(target, error, 'Could not resend OTP.'),
      );
      throw error;
    }
  }

  return (
    <AuthPageFrame>
      <AuthCard
        description="Enter the 6 digit code sent to your email or phone."
        title="Verify OTP"
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
              OTP challenge is missing. Resend OTP to continue.
            </Alert>
          ) : null}

          {serverError ? <Alert variant="error">{serverError}</Alert> : null}

          {deliveryNotice ? <Alert variant="info">{deliveryNotice}</Alert> : null}

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

          <AuthSubmit isSubmitting={isSubmitting}>Verify OTP</AuthSubmit>

          <ResendOtpButton disabled={!target} onResend={handleResend} />
        </form>
      </AuthCard>
    </AuthPageFrame>
  );
}
