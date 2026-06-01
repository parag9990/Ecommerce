import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Input } from '../../../components/ui/input';
import { PasswordInput } from '../../../components/ui/password-input';
import { routePaths } from '../../../routes/route-paths';
import { sendOtp, signup } from '../api/auth.api';
import { AuthCard } from '../components/auth-card';
import { AuthField } from '../components/auth-field';
import { AuthPageFrame } from '../components/auth-page-frame';
import { AuthSubmit } from '../components/auth-submit';
import { PasswordStrength } from '../components/password-strength';
import { signupSchema, type SignupFormValues } from '../schemas';
import type { SignupRequest } from '../types';
import { toTrimmedValue } from '../utils';

const fieldIds = {
  acceptTerms: 'signup-accept-terms',
  confirmPassword: 'signup-confirm-password',
  email: 'signup-email',
  fullName: 'signup-full-name',
  password: 'signup-password',
  phone: 'signup-phone',
} as const;

export function SignupPage() {
  const navigate = useNavigate();
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<SignupFormValues>({
    defaultValues: {
      acceptTerms: false,
      confirmPassword: '',
      email: '',
      full_name: '',
      password: '',
      phone: '',
    },
    resolver: zodResolver(signupSchema),
  });

  const [password, setPassword] = useState('');
  const passwordRegistration = register('password');

  async function onSubmit(values: SignupFormValues) {
    setServerError(null);

    const email = toTrimmedValue(values.email).toLowerCase();
    const phone = toTrimmedValue(values.phone);
    const payload: SignupRequest = {
      email,
      full_name: toTrimmedValue(values.full_name),
      password: values.password,
      role: 'buyer',
    };

    if (phone) {
      payload.phone = phone;
    }

    try {
      await signup(payload);

      const challenge = await sendOtp({
        channel: 'email',
        purpose: 'signup',
        target: email,
      });
      const params = new URLSearchParams({
        challenge_id: challenge.challenge_id,
        purpose: 'signup',
        target: email,
      });

      await navigate(`${routePaths.otp}?${params.toString()}`, {
        replace: true,
      });
    } catch (error) {
      setServerError(error instanceof Error ? error.message : 'Signup failed.');
    }
  }

  return (
    <AuthPageFrame>
      <AuthCard
        description="Create a buyer account with secure credentials."
        title="Create your account"
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
            error={errors.full_name?.message}
            errorId={`${fieldIds.fullName}-error`}
            htmlFor={fieldIds.fullName}
            label="Full name"
          >
            <Input
              aria-describedby={
                errors.full_name ? `${fieldIds.fullName}-error` : undefined
              }
              aria-invalid={Boolean(errors.full_name)}
              autoComplete="name"
              id={fieldIds.fullName}
              invalid={Boolean(errors.full_name)}
              type="text"
              {...register('full_name')}
            />
          </AuthField>

          <AuthField
            error={errors.email?.message}
            errorId={`${fieldIds.email}-error`}
            htmlFor={fieldIds.email}
            label="Email"
          >
            <Input
              aria-describedby={
                errors.email ? `${fieldIds.email}-error` : undefined
              }
              aria-invalid={Boolean(errors.email)}
              autoComplete="email"
              id={fieldIds.email}
              invalid={Boolean(errors.email)}
              type="email"
              {...register('email')}
            />
          </AuthField>

          <AuthField
            description="Optional, but useful for delivery and account recovery."
            error={errors.phone?.message}
            errorId={`${fieldIds.phone}-error`}
            htmlFor={fieldIds.phone}
            label="Phone"
          >
            <Input
              aria-describedby={
                errors.phone ? `${fieldIds.phone}-error` : undefined
              }
              aria-invalid={Boolean(errors.phone)}
              autoComplete="tel"
              id={fieldIds.phone}
              invalid={Boolean(errors.phone)}
              inputMode="tel"
              type="tel"
              {...register('phone')}
            />
          </AuthField>

          <AuthField
            error={errors.password?.message}
            errorId={`${fieldIds.password}-error`}
            htmlFor={fieldIds.password}
            label="Password"
          >
            <PasswordInput
              aria-describedby={
                errors.password ? `${fieldIds.password}-error` : undefined
              }
              aria-invalid={Boolean(errors.password)}
              autoComplete="new-password"
              id={fieldIds.password}
              invalid={Boolean(errors.password)}
              {...passwordRegistration}
              onChange={(event) => {
                setPassword(event.target.value);
                void passwordRegistration.onChange(event);
              }}
            />
          </AuthField>

          <PasswordStrength password={password} />

          <AuthField
            error={errors.confirmPassword?.message}
            errorId={`${fieldIds.confirmPassword}-error`}
            htmlFor={fieldIds.confirmPassword}
            label="Confirm password"
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

          <div>
            <label className="flex items-start gap-2 text-sm leading-6 text-slate-700">
              <input
                aria-describedby={
                  errors.acceptTerms
                    ? `${fieldIds.acceptTerms}-error`
                    : undefined
                }
                aria-invalid={Boolean(errors.acceptTerms)}
                className="mt-1 h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                id={fieldIds.acceptTerms}
                type="checkbox"
                {...register('acceptTerms')}
              />
              <span>I agree to the terms and privacy policy.</span>
            </label>
            {errors.acceptTerms?.message ? (
              <p
                className="mt-1 text-sm leading-5 text-red-600"
                id={`${fieldIds.acceptTerms}-error`}
              >
                {errors.acceptTerms.message}
              </p>
            ) : null}
          </div>

          <AuthSubmit isSubmitting={isSubmitting}>Create account</AuthSubmit>
        </form>

        <p className="mt-6 text-center text-sm text-slate-600">
          Already have an account?{' '}
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
