import type {
  Control,
  FieldErrors,
  UseFormRegister,
  UseFormSetValue,
} from "react-hook-form";
import { useFieldArray, useWatch } from "react-hook-form";
import { Plus, Trash2 } from "lucide-react";

import { IconButton } from "../../../components/ui/icon-button";
import type { ProductFormValues } from "../utils/product-validation";
import { createEmptyVariant } from "../utils/product-validation";
import { AttributeListEditor } from "./attribute-list-editor";

type VariantEditorProps = {
  control: Control<ProductFormValues>;
  register: UseFormRegister<ProductFormValues>;
  setValue: UseFormSetValue<ProductFormValues>;
  errors: FieldErrors<ProductFormValues>;
};

type VariantAttributesProps = {
  control: Control<ProductFormValues>;
  setValue: UseFormSetValue<ProductFormValues>;
  index: number;
};

function VariantAttributes({ control, setValue, index }: VariantAttributesProps) {
  const attributes =
    useWatch({
      control,
      name: `variants.${index}.attributes`,
    }) ?? {};

  return (
    <AttributeListEditor
      label="Attributes"
      compact
      value={attributes}
      onChange={(nextAttributes) =>
        setValue(`variants.${index}.attributes`, nextAttributes, {
          shouldDirty: true,
          shouldValidate: true,
        })
      }
    />
  );
}

export function VariantEditor({ control, register, setValue, errors }: VariantEditorProps) {
  const { fields, append, remove } = useFieldArray({
    control,
    name: "variants",
  });
  const variantsError = errors.variants?.message;

  return (
    <section className="space-y-3 border-t border-slate-100 pt-4">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold text-slate-950">Variants</h2>
          {variantsError ? <p className="text-xs text-rose-600">{variantsError}</p> : null}
        </div>
        <button
          type="button"
          onClick={() => append(createEmptyVariant())}
          className="inline-flex h-8 items-center gap-1.5 rounded-md border border-slate-200 bg-white px-2.5 text-xs font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        >
          <Plus className="h-3.5 w-3.5" aria-hidden="true" />
          Add variant
        </button>
      </div>

      <div className="space-y-3">
        {fields.map((field, index) => {
          const variantErrors = errors.variants?.[index];

          return (
            <div
              key={field.id}
              className="space-y-3 rounded-md border border-slate-200 bg-slate-50 p-3"
            >
              <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(8rem,10rem)_minmax(8rem,10rem)_auto]">
                <label className="space-y-1">
                  <span className="text-xs font-medium text-slate-600">SKU</span>
                  <input
                    {...register(`variants.${index}.sku`)}
                    className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                    placeholder="TSHIRT-BLK-M"
                    aria-invalid={Boolean(variantErrors?.sku)}
                  />
                  {variantErrors?.sku?.message ? (
                    <p className="text-xs text-rose-600">{variantErrors.sku.message}</p>
                  ) : null}
                </label>

                <label className="space-y-1">
                  <span className="text-xs font-medium text-slate-600">Price</span>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    {...register(`variants.${index}.price.amount`, {
                      valueAsNumber: true,
                    })}
                    className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                    placeholder="999"
                    aria-invalid={Boolean(variantErrors?.price?.amount)}
                  />
                  {variantErrors?.price?.amount?.message ? (
                    <p className="text-xs text-rose-600">
                      {variantErrors.price.amount.message}
                    </p>
                  ) : null}
                </label>

                <label className="space-y-1">
                  <span className="text-xs font-medium text-slate-600">Stock</span>
                  <input
                    type="number"
                    min="0"
                    step="1"
                    {...register(`variants.${index}.stock_quantity`, {
                      valueAsNumber: true,
                    })}
                    className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                    placeholder="25"
                    aria-invalid={Boolean(variantErrors?.stock_quantity)}
                  />
                  {variantErrors?.stock_quantity?.message ? (
                    <p className="text-xs text-rose-600">
                      {variantErrors.stock_quantity.message}
                    </p>
                  ) : null}
                </label>

                <div className="flex items-end justify-end">
                  <IconButton
                    label="Remove variant"
                    disabled={fields.length === 1}
                    onClick={() => remove(index)}
                    className="h-10 w-10 text-rose-600 hover:text-rose-700 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <Trash2 className="h-4 w-4" aria-hidden="true" />
                  </IconButton>
                </div>
              </div>

              <VariantAttributes control={control} setValue={setValue} index={index} />
            </div>
          );
        })}
      </div>
    </section>
  );
}
