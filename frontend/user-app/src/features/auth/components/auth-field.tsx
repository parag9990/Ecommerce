import type { ReactNode } from 'react';

type AuthFieldProps = {
  children: ReactNode;
  description?: string | undefined;
  error?: string | undefined;
  errorId?: string | undefined;
  htmlFor: string;
  label: string;
};

export function AuthField({
  children,
  description,
  error,
  errorId,
  htmlFor,
  label,
}: AuthFieldProps) {
  return (
    <div className="space-y-1.5">
      <label className="text-sm font-medium text-slate-800" htmlFor={htmlFor}>
        {label}
      </label>
      {description ? (
        <p className="text-xs leading-5 text-slate-500">{description}</p>
      ) : null}
      {children}
      {error ? (
        <p className="text-sm leading-5 text-red-600" id={errorId}>
          {error}
        </p>
      ) : null}
    </div>
  );
}
