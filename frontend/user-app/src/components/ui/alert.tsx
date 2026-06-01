import type { ReactNode } from 'react';

type AlertVariant = 'error' | 'info' | 'success';

type AlertProps = {
  children: ReactNode;
  title?: string;
  variant?: AlertVariant;
};

const variantClasses: Record<AlertVariant, string> = {
  error: 'border-red-200 bg-red-50 text-red-700',
  info: 'border-blue-200 bg-blue-50 text-blue-700',
  success: 'border-emerald-200 bg-emerald-50 text-emerald-700',
};

export function Alert({ children, title, variant = 'info' }: AlertProps) {
  const role = variant === 'error' ? 'alert' : 'status';

  return (
    <div
      aria-live={variant === 'error' ? 'assertive' : 'polite'}
      className={[
        'rounded-md border px-3 py-2 text-sm leading-6',
        variantClasses[variant],
      ].join(' ')}
      role={role}
    >
      {title ? <p className="font-semibold">{title}</p> : null}
      <div>{children}</div>
    </div>
  );
}
