import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Input } from '../../../components/ui/input';
import { routePaths } from '../../../routes/route-paths';
import { forgotPassword } from '../api/auth.api';
import { AuthCard } from '../components/auth-card';
import { AuthField } from '../components/auth-field';
import { AuthPageFrame } from '../components/auth-page-frame';
import { AuthSubmit } from '../components/auth-submit';
import {
  forgotPasswordSchema,
  type ForgotPasswordFormValues,
} from '../schemas';
import { otpDeliveryErrorMessage, toTrimmedValue } from '../utils';

const fieldIds = {
  identifier: 'forgot-password-identifier',
} as const;

export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<ForgotPasswordFormValues>({
    defaultValues: {
      identifier: '',
    },
    resolver: zodResolver(forgotPasswordSchema),
  });

  async function onSubmit(values: ForgotPasswordFormValues) {
    setServerError(null);

    try {
      const identifier = toTrimmedValue(values.identifier);
      const challenge = await forgotPassword({ identifier });
      const params = new URLSearchParams({
        challenge_id: challenge.challenge_id,
        target: identifier,
      });

      await navigate(`${routePaths.resetPassword}?${params.toString()}`);
    } catch (error) {
      setServerError(
        otpDeliveryErrorMessage(
          toTrimmedValue(values.identifier),
          error,
          'Could not start password reset.',
        ),
      );
    }
  }

  return (
    <AuthPageFrame>
      <AuthCard
        description="Enter your email or phone and we will send a reset OTP."
        title="Forgot password"
      >
        <form
          className="space-y-4"
          noValidate
          onSubmit={(event) => {
            void handleSubmit(onSubmit)(event);
          }}
        >
          {serverError ? <Alert variant="error">{serverError}</Alert> : null}

          <AuthField
            error={errors.identifier?.message}
            errorId={`${fieldIds.identifier}-error`}
            htmlFor={fieldIds.identifier}
            label="Email or phone"
          >
            <Input
              aria-describedby={
                errors.identifier ? `${fieldIds.identifier}-error` : undefined
              }
              aria-invalid={Boolean(errors.identifier)}
              autoComplete="username"
              id={fieldIds.identifier}
              invalid={Boolean(errors.identifier)}
              type="text"
              {...register('identifier')}
            />
          </AuthField>

          <AuthSubmit isSubmitting={isSubmitting}>Send reset OTP</AuthSubmit>
        </form>

        <p className="mt-6 text-center text-sm text-slate-600">
          Remember your password?{' '}
          <Link
            className="font-medium text-blue-700 hover:text-blue-800"
            to={routePaths.login}
          >
            Login
          </Link>
        </p>
      </AuthCard>
    </AuthPageFrame>
  );
}
