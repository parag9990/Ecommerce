import { Star } from 'lucide-react';

type RatingProps = {
  value?: number | undefined;
};

export function Rating({ value }: RatingProps) {
  if (value === undefined || value <= 0) {
    return <span className="text-xs text-slate-500">No ratings yet</span>;
  }

  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-1 text-xs font-medium text-amber-700">
      <Star aria-hidden="true" className="h-3.5 w-3.5 fill-current" />
      {value.toFixed(1)}
    </span>
  );
}
