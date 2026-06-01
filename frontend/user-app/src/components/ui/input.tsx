import { forwardRef } from 'react';
import type { InputHTMLAttributes } from 'react';

type InputProps = InputHTMLAttributes<HTMLInputElement> & {
  invalid?: boolean;
};

export const inputClasses =
  'w-full rounded-md border bg-white px-3 py-2.5 text-sm text-slate-950 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500';

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className = '', invalid = false, ...props }, ref) => (
    <input
      className={[
        inputClasses,
        invalid ? 'border-red-300' : 'border-slate-300',
        className,
      ].join(' ')}
      ref={ref}
      {...props}
    />
  ),
);

Input.displayName = 'Input';
