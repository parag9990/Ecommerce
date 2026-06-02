import { zodResolver } from "@hookform/resolvers/zod";
import { Save } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";

import type { Product, ProductInput } from "../types";
import { toProductFormValues, toProductInput } from "../utils/product-mappers";
import type { ProductFormValues } from "../utils/product-validation";
import { productFormSchema } from "../utils/product-validation";
import { AttributeListEditor } from "./attribute-list-editor";
import { CategorySelector } from "./category-selector";
import { ImageUploader } from "./image-uploader";
import { PublishChecklist } from "./publish-checklist";
import { VariantEditor } from "./variant-editor";

type ProductFormProps = {
  mode: "create" | "edit";
  product?: Product;
  saving: boolean;
  onSubmit: (input: ProductInput) => Promise<void>;
};

function FormError({ message }: { message?: string }) {
  if (!message) {
    return null;
  }

  return <p className="text-xs text-rose-600">{message}</p>;
}

export function ProductForm({ mode, product, saving, onSubmit }: ProductFormProps) {
  const form = useForm<ProductFormValues>({
    resolver: zodResolver(productFormSchema),
    defaultValues: toProductFormValues(product),
    mode: "onBlur",
  });

  const title = form.watch("title");
  const categoryId = form.watch("category_id");
  const attributes = form.watch("attributes");
  const images = form.watch("images");
  const variants = form.watch("variants");
  const { reset } = form;

  useEffect(() => {
    reset(toProductFormValues(product));
  }, [product, reset]);

  return (
    <form
      onSubmit={form.handleSubmit((values) => onSubmit(toProductInput(values)))}
      className="space-y-4 rounded-md border border-slate-200 bg-white p-4 shadow-sm"
    >
      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Title</span>
              <input
                {...form.register("title")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="Cotton T-shirt"
                aria-invalid={Boolean(form.formState.errors.title)}
              />
              <FormError message={form.formState.errors.title?.message} />
            </label>

            <label className="space-y-1">
              <span className="text-sm font-medium text-slate-700">Brand</span>
              <input
                {...form.register("brand")}
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
                placeholder="Brand name"
              />
              <FormError message={form.formState.errors.brand?.message} />
            </label>
          </div>

          <CategorySelector
            value={categoryId}
            onChange={(nextCategoryId) =>
              form.setValue("category_id", nextCategoryId, {
                shouldDirty: true,
                shouldValidate: true,
              })
            }
            error={form.formState.errors.category_id?.message}
          />

          <label className="space-y-1">
            <span className="text-sm font-medium text-slate-700">Description</span>
            <textarea
              {...form.register("description")}
              className="min-h-28 w-full resize-y rounded-md border border-slate-300 px-3 py-2 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              placeholder="Product details, material, care instructions"
            />
            <FormError message={form.formState.errors.description?.message} />
          </label>
        </div>

        <PublishChecklist
          hasTitle={title.trim().length >= 3}
          hasCategory={Boolean(categoryId)}
          hasVariant={variants.some((variant) => variant.sku.trim() && variant.price.amount > 0)}
          hasImage={images.length > 0}
        />
      </div>

      <section className="border-t border-slate-100 pt-4">
        <AttributeListEditor
          label="Product attributes"
          value={attributes}
          onChange={(nextAttributes) =>
            form.setValue("attributes", nextAttributes, {
              shouldDirty: true,
              shouldValidate: true,
            })
          }
        />
      </section>

      <ImageUploader
        images={images}
        error={form.formState.errors.images?.message}
        onChange={(nextImages) =>
          form.setValue("images", nextImages, {
            shouldDirty: true,
            shouldValidate: true,
          })
        }
      />

      <VariantEditor
        control={form.control}
        register={form.register}
        setValue={form.setValue}
        errors={form.formState.errors}
      />

      <div className="flex items-center justify-end gap-2 border-t border-slate-100 pt-4">
        <button
          type="submit"
          disabled={saving}
          className="inline-flex h-10 items-center gap-2 rounded-md bg-slate-950 px-4 text-sm font-medium text-white transition hover:bg-slate-800 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-950 disabled:cursor-not-allowed disabled:opacity-60"
        >
          <Save className="h-4 w-4" aria-hidden="true" />
          {saving ? "Saving..." : mode === "create" ? "Save draft" : "Save changes"}
        </button>
      </div>
    </form>
  );
}
