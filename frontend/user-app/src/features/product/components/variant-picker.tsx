import type { ProductVariant } from '../types';
import { Price } from './price';

type VariantPickerProps = {
  onChange: (sku: string) => void;
  selectedSku?: string | undefined;
  variants: ProductVariant[];
};

function formatVariantAttributes(variant: ProductVariant) {
  if (!variant.attributes) {
    return null;
  }

  const attributes = Object.entries(variant.attributes).filter(([, value]) => {
    return ['boolean', 'number', 'string'].includes(typeof value);
  });

  if (attributes.length === 0) {
    return null;
  }

  return attributes
    .map(([key, value]) => `${key}: ${String(value)}`)
    .join(' · ');
}

export function VariantPicker({
  onChange,
  selectedSku,
  variants,
}: VariantPickerProps) {
  if (variants.length === 0) {
    return (
      <p className="rounded-md border border-slate-200 bg-slate-50 p-3 text-sm text-slate-600">
        Variant information is unavailable.
      </p>
    );
  }

  return (
    <fieldset className="space-y-3">
      <legend className="font-medium text-slate-950">Choose variant</legend>

      <div className="grid gap-2 sm:grid-cols-2">
        {variants.map((variant) => {
          const isSelected = selectedSku === variant.sku;
          const stock = variant.stock_quantity;
          const attributeLabel = formatVariantAttributes(variant);

          return (
            <label
              className={[
                'rounded-md border p-3 transition',
                isSelected
                  ? 'border-slate-950 bg-slate-50'
                  : 'border-slate-200 bg-white hover:border-slate-400',
              ].join(' ')}
              key={variant.sku}
            >
              <input
                checked={isSelected}
                className="sr-only"
                name="variant"
                onChange={() => {
                  onChange(variant.sku);
                }}
                type="radio"
                value={variant.sku}
              />
              <span className="block text-sm font-medium text-slate-950">
                {variant.sku}
              </span>
              {attributeLabel ? (
                <span className="mt-1 block text-xs text-slate-500">
                  {attributeLabel}
                </span>
              ) : null}
              <span className="mt-2 block text-sm">
                <Price value={variant.price} />
              </span>
              <span
                className={
                  stock === undefined
                    ? 'mt-1 block text-xs text-slate-500'
                    : stock > 0
                      ? 'mt-1 block text-xs text-emerald-700'
                      : 'mt-1 block text-xs text-red-700'
                }
              >
                {stock === undefined
                  ? 'Stock unavailable'
                  : stock > 0
                    ? `${stock} in stock`
                    : 'Out of stock'}
              </span>
            </label>
          );
        })}
      </div>
    </fieldset>
  );
}
