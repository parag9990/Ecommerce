import { Minus, Plus } from 'lucide-react';

import { Button } from '../../../components/ui/button';

type QuantityStepperProps = {
  disabled?: boolean | undefined;
  onChange: (nextQuantity: number) => void;
  value: number;
};

export function QuantityStepper({
  disabled = false,
  onChange,
  value,
}: QuantityStepperProps) {
  return (
    <div className="inline-grid h-10 grid-cols-[2.5rem_3rem_2.5rem] overflow-hidden rounded-md border border-slate-300 bg-white">
      <Button
        aria-label="Decrease quantity"
        className="h-10 rounded-none border-0 px-0"
        disabled={disabled || value <= 1}
        onClick={() => {
          onChange(value - 1);
        }}
        type="button"
        variant="ghost"
      >
        <Minus aria-hidden="true" className="h-4 w-4" />
      </Button>
      <output
        aria-label="Quantity"
        className="flex items-center justify-center border-x border-slate-200 text-sm font-semibold text-slate-950"
      >
        {value}
      </output>
      <Button
        aria-label="Increase quantity"
        className="h-10 rounded-none border-0 px-0"
        disabled={disabled}
        onClick={() => {
          onChange(value + 1);
        }}
        type="button"
        variant="ghost"
      >
        <Plus aria-hidden="true" className="h-4 w-4" />
      </Button>
    </div>
  );
}
