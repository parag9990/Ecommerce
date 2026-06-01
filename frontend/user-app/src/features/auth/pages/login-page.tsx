import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';

import { Alert } from '../../../components/ui/alert';
import { Input } from '../../../components/ui/input';
import { PasswordInput } from '../../../components/ui/password-input';
import { routePaths } from '../../../routes/route-paths';
import { AuthCard } from '../components/auth-card';
import { AuthField } from '../components/auth-field';
import { AuthPageFrame } from '../components/auth-page-frame';
import { AuthSubmit } from '../components/auth-submit';
import { useLoginMutation } from '../hooks/use-login-mutation';
import { loginSchema, type LoginFormValues } from '../schemas';
import { getDeviceInfo, safeRedirectPath, toTrimmedValue } from '../utils';

const fieldIds = {
  identifier: 'login-identifier',
  password: 'login-password',
} as const;

function readLoginNotice(searchParams: URLSearchParams): string | null {
  if (searchParams.get('verified') === '1') {
    return 'OTP verified. Please login to continue.';
  }

  if (searchParams.get('passwordReset') === '1') {
    return 'Password reset successful. Please login with your new password.';
  }

  return null;
}

export function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [serverError, setServerError] = useState<string | null>(null);
  const loginMutation = useLoginMutation();
  const loginNotice = readLoginNotice(searchParams);

  const {
    formState: { errors, isSubmitting },
    handleSubmit,
    register,
  } = useForm<LoginFormValues>({
    defaultValues: {
      identifier: '',
      password: '',
    },
    resolver: zodResolver(loginSchema),
  });

  async function onSubmit(values: LoginFormValues) {
    setServerError(null);

    try {
      await loginMutation.mutateAsync({
        device: getDeviceInfo(),
        identifier: toTrimmedValue(values.identifier),
        password: values.password,
      });

      await navigate(safeRedirectPath(searchParams.get('redirect')), {
        replace: true,
      });
    } catch (error) {
      setServerError(error instanceof Error ? error.message : 'Login failed.');
    }
  }

  return (
    <AuthPageFrame>
      <AuthCard
        description="Login with your email or phone to continue shopping."
        title="Welcome back"
      >
        <form
          className="space-y-4"
          noValidate
          onSubmit={(event) => {
            void handleSubmit(onSubmit)(event);
          }}
        >
          {loginNotice ? <Alert variant="success">{loginNotice}</Alert> : null}

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
              autoComplete="current-password"
              id={fieldIds.password}
              invalid={Boolean(errors.password)}
              {...register('password')}
            />
          </AuthField>

          <div className="flex justify-end">
            <Link
              className="rounded text-sm font-medium text-blue-700 hover:text-blue-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
              to={routePaths.forgotPassword}
            >
              Forgot password?
            </Link>
          </div>

          <AuthSubmit isSubmitting={isSubmitting}>Login</AuthSubmit>
        </form>

        <p className="mt-6 text-center text-sm text-slate-600">
          New user?{' '}
          <Link
            className="font-medium text-blue-700 hover:text-blue-800"
            to={routePaths.signup}
          >
            Create account
          </Link>
        </p>
      </AuthCard>
    </AuthPageFrame>
  );
}
