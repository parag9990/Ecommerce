import type { ReactNode } from 'react';

type AuthPageFrameProps = {
  children: ReactNode;
};

export function AuthPageFrame({ children }: AuthPageFrameProps) {
  return (
    <div className="flex min-h-[calc(100vh-9rem)] items-start justify-center py-4 sm:items-center sm:py-8">
      {children}
    </div>
  );
}
