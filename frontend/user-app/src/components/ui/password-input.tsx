import { Eye, EyeOff } from 'lucide-react';
import { forwardRef, useState } from 'react';
import type { InputHTMLAttributes } from 'react';

import { Input } from './input';

type PasswordInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> & {
  invalid?: boolean;
};

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ className = '', disabled, ...props }, ref) => {
    const [isVisible, setIsVisible] = useState(false);
    const Icon = isVisible ? EyeOff : Eye;

    return (
      <div className="relative">
        <Input
          className={['pr-11', className].join(' ')}
          disabled={disabled}
          ref={ref}
          type={isVisible ? 'text' : 'password'}
          {...props}
        />
        <button
          aria-label={isVisible ? 'Hide password' : 'Show password'}
          className={[
            'absolute right-2 top-1/2 inline-flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-md text-slate-500 transition',
            'hover:bg-slate-100 hover:text-slate-800',
            'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600',
            'disabled:cursor-not-allowed disabled:text-slate-300',
          ].join(' ')}
          disabled={disabled}
          onClick={() => {
            setIsVisible((current) => !current);
          }}
          type="button"
        >
          <Icon aria-hidden="true" className="h-4 w-4" />
        </button>
      </div>
    );
  },
);

PasswordInput.displayName = 'PasswordInput';
