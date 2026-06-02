import { describe, expect, it } from "vitest";

import { productFormSchema } from "./product-validation";

const validProduct = {
  title: "Cotton Shirt",
  description: "Soft cotton shirt",
  brand: "North",
  category_id: "cat-shirts",
  attributes: {
    material: "cotton",
  },
  images: ["https://cdn.example.com/shirt.jpg"],
  variants: [
    {
      sku: "SHIRT-BLK-M",
      attributes: {
        size: "M",
      },
      price: {
        amount: 999,
        currency: "INR",
      },
      stock_quantity: 12,
    },
  ],
};

describe("productFormSchema", () => {
  it("accepts a valid product payload", () => {
    expect(productFormSchema.safeParse(validProduct).success).toBe(true);
  });

  it("rejects missing title and category", () => {
    const result = productFormSchema.safeParse({
      ...validProduct,
      title: "",
      category_id: "",
    });

    expect(result.success).toBe(false);
  });

  it("rejects negative stock", () => {
    const result = productFormSchema.safeParse({
      ...validProduct,
      variants: [
        {
          ...validProduct.variants[0],
          stock_quantity: -1,
        },
      ],
    });

    expect(result.success).toBe(false);
  });

  it("rejects invalid image URLs", () => {
    const result = productFormSchema.safeParse({
      ...validProduct,
      images: ["not-a-url"],
    });

    expect(result.success).toBe(false);
  });
});
