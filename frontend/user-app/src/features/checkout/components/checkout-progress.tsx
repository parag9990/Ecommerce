import { Check } from 'lucide-react';

type CheckoutProgressProps = {
  activeStep: 'address' | 'review' | 'payment';
};

const steps = [
  { id: 'address', label: 'Address' },
  { id: 'review', label: 'Review' },
  { id: 'payment', label: 'Payment' },
] as const;

export function CheckoutProgress({ activeStep }: CheckoutProgressProps) {
  const activeIndex = steps.findIndex((step) => step.id === activeStep);

  return (
    <ol className="grid grid-cols-3 gap-2" aria-label="Checkout progress">
      {steps.map((step, index) => {
        const isComplete = index < activeIndex;
        const isActive = index === activeIndex;

        return (
          <li
            className={[
              'flex min-w-0 items-center gap-2 rounded-md border px-3 py-2 text-sm',
              isActive
                ? 'border-blue-600 bg-blue-50 font-semibold text-blue-700'
                : 'border-slate-200 bg-white text-slate-600',
            ].join(' ')}
            key={step.id}
          >
            <span
              className={[
                'flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-xs',
                isComplete || isActive
                  ? 'bg-blue-600 text-white'
                  : 'bg-slate-200 text-slate-600',
              ].join(' ')}
            >
              {isComplete ? (
                <Check aria-hidden="true" className="h-3.5 w-3.5" />
              ) : (
                index + 1
              )}
            </span>
            <span className="truncate">{step.label}</span>
          </li>
        );
      })}
    </ol>
  );
}
