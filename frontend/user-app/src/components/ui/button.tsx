import type { ButtonHTMLAttributes, ReactNode } from 'react';

type ButtonVariant = 'ghost' | 'primary' | 'secondary';

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  fullWidth?: boolean;
  variant?: ButtonVariant;
};

const variantClasses: Record<ButtonVariant, string> = {
  ghost:
    'text-blue-700 hover:bg-blue-50 hover:text-blue-800 disabled:text-slate-400',
  primary:
    'bg-blue-600 text-white hover:bg-blue-700 focus-visible:outline-blue-600 disabled:bg-slate-400',
  secondary:
    'border border-slate-300 bg-white text-slate-700 hover:bg-slate-50 hover:text-slate-950 disabled:text-slate-400',
};

export function Button({
  children,
  className = '',
  fullWidth = false,
  type = 'button',
  variant = 'primary',
  ...props
}: ButtonProps) {
  return (
    <button
      className={[
        'inline-flex h-11 items-center justify-center rounded-md px-4 text-sm font-semibold transition',
        'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2',
        'disabled:cursor-not-allowed',
        fullWidth ? 'w-full' : '',
        variantClasses[variant],
        className,
      ].join(' ')}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
}
