import { z } from "zod";

const attributeRecordSchema = z.record(z.string(), z.string());

export const productVariantSchema = z.object({
  sku: z.string().trim().min(2, "Enter a SKU with at least 2 characters."),
  attributes: attributeRecordSchema,
  price: z.object({
    amount: z.number().positive("Enter a price greater than 0."),
    currency: z.literal("INR"),
  }),
  stock_quantity: z
    .number()
    .int("Stock must be a whole number.")
    .min(0, "Stock cannot be negative."),
});

export const productFormSchema = z.object({
  title: z.string().trim().min(3, "Enter a product title with at least 3 characters."),
  description: z.string().trim().optional(),
  brand: z.string().trim().optional(),
  category_id: z.string().min(1, "Select a category."),
  attributes: attributeRecordSchema,
  images: z.array(z.string().trim().url("Enter a valid image URL.")),
  variants: z.array(productVariantSchema).min(1, "Add at least one variant."),
});

export type ProductFormValues = z.infer<typeof productFormSchema>;

export const emptyVariant: ProductFormValues["variants"][number] = {
  sku: "",
  attributes: {},
  price: {
    amount: 0,
    currency: "INR",
  },
  stock_quantity: 0,
};

export function createEmptyVariant(): ProductFormValues["variants"][number] {
  return {
    sku: "",
    attributes: {},
    price: {
      amount: 0,
      currency: "INR",
    },
    stock_quantity: 0,
  };
}
