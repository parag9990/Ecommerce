import type { ButtonHTMLAttributes, ReactNode } from 'react';

type IconButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  label: string;
};

export function IconButton({
  children,
  className = '',
  label,
  type = 'button',
  ...props
}: IconButtonProps) {
  return (
    <button
      aria-label={label}
      className={[
        'inline-flex h-10 w-10 items-center justify-center rounded-md text-slate-700 transition',
        'hover:bg-slate-100 hover:text-slate-950',
        'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600',
        className,
      ].join(' ')}
      type={type}
      {...props}
    >
      {children}
    </button>
  );
}
