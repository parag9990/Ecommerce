import { describe, expect, it } from "vitest";

import type { ProductFormValues } from "./product-validation";
import { normalizeProduct, toProductInput } from "./product-mappers";

describe("product mappers", () => {
  it("trims form values and removes empty attributes", () => {
    const values: ProductFormValues = {
      title: "  Linen Shirt  ",
      description: "  Summer fit  ",
      brand: "  North  ",
      category_id: "cat-shirts",
      attributes: {
        " material ": " linen ",
        empty: "",
      },
      images: [" https://cdn.example.com/linen.jpg "],
      variants: [
        {
          sku: " LINEN-WHT-M ",
          attributes: {
            " size ": " M ",
          },
          price: {
            amount: 1499,
            currency: "INR",
          },
          stock_quantity: 5,
        },
      ],
    };

    expect(toProductInput(values)).toEqual({
      title: "Linen Shirt",
      description: "Summer fit",
      brand: "North",
      category_id: "cat-shirts",
      attributes: {
        material: "linen",
      },
      images: ["https://cdn.example.com/linen.jpg"],
      variants: [
        {
          sku: "LINEN-WHT-M",
          attributes: {
            size: "M",
          },
          price: {
            amount: 1499,
            currency: "INR",
          },
          stock_quantity: 5,
        },
      ],
    });
  });

  it("normalizes product API responses defensively", () => {
    expect(
      normalizeProduct({
        product_id: "product-1",
        seller_id: "seller-1",
        title: "Mug",
        category_id: "cat-home",
        status: "published",
        attributes: {
          capacity: 350,
        },
        images: ["https://cdn.example.com/mug.jpg"],
        variants: [
          {
            sku: "MUG-BLK",
            price: {
              amount: "499",
            },
            stock_quantity: "9",
          },
        ],
      }),
    ).toMatchObject({
      product_id: "product-1",
      seller_id: "seller-1",
      status: "published",
      attributes: {
        capacity: "350",
      },
      images: ["https://cdn.example.com/mug.jpg"],
      variants: [
        {
          sku: "MUG-BLK",
          price: {
            amount: 499,
            currency: "INR",
          },
          stock_quantity: 9,
        },
      ],
    });
  });

  it("normalizes object-shaped product images into ordered URLs", () => {
    expect(
      normalizeProduct({
        product_id: "product-1",
        seller_id: "seller-1",
        title: "Marble Look Tray",
        category_id: "cat-home-decor",
        status: "published",
        images: [
          {
            url: " https://cdn.example.com/tray-side.jpg ",
            position: 2,
            status: "active",
          },
          {
            url: "https://cdn.example.com/tray-primary.jpg",
            is_primary: true,
            position: 3,
            status: "active",
          },
          {
            url: "https://cdn.example.com/tray-archived.jpg",
            position: 1,
            status: "inactive",
          },
          "https://cdn.example.com/tray-string.jpg",
        ],
        variants: [],
      }).images,
    ).toEqual([
      "https://cdn.example.com/tray-primary.jpg",
      "https://cdn.example.com/tray-side.jpg",
      "https://cdn.example.com/tray-string.jpg",
    ]);
  });
});
