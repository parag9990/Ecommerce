import { Percent, Tag } from "lucide-react";

import { cn } from "../../../lib/cn";
import type { DiscountType } from "../types";

type DiscountTypeSelectorProps = {
  value: DiscountType;
  onChange: (value: DiscountType) => void;
};

const options: Array<{ value: DiscountType; label: string; icon: typeof Percent }> = [
  { value: "percentage", label: "Percentage", icon: Percent },
  { value: "fixed", label: "Fixed amount", icon: Tag },
];

export function DiscountTypeSelector({ value, onChange }: DiscountTypeSelectorProps) {
  return (
    <div className="grid grid-cols-2 gap-2" role="group" aria-label="Discount type">
      {options.map((option) => {
        const Icon = option.icon;

        return (
          <button
            key={option.value}
            type="button"
            onClick={() => onChange(option.value)}
            className={cn(
              "inline-flex h-10 items-center justify-center gap-2 rounded-md border px-3 text-sm font-medium transition",
              "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600",
              option.value === value
                ? "border-blue-600 bg-blue-50 text-blue-700"
                : "border-slate-200 bg-white text-slate-600 hover:bg-slate-50 hover:text-slate-950",
            )}
          >
            <Icon className="h-4 w-4" aria-hidden="true" />
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
