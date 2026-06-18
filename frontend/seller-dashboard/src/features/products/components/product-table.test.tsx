import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import type { Product } from "../types";
import { ProductTable } from "./product-table";

const product: Product = {
  product_id: "product-1",
  seller_id: "seller-1",
  title: "Cotton Shirt",
  description: "Soft cotton shirt",
  brand: "North",
  category_id: "cat-shirts",
  status: "draft",
  attributes: {},
  images: [],
  variants: [
    {
      sku: "SHIRT-BLK-M",
      attributes: {},
      price: {
        amount: 999,
        currency: "INR",
      },
      stock_quantity: 12,
    },
  ],
};

describe("ProductTable", () => {
  it("renders the empty state", () => {
    render(
      <MemoryRouter>
        <ProductTable
          products={[]}
          total={0}
          page={1}
          pageSize={20}
          categories={[]}
          onPageChange={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("No products")).toBeTruthy();
  });

  it("renders product rows with status and category names", () => {
    render(
      <MemoryRouter>
        <ProductTable
          products={[product]}
          total={1}
          page={1}
          pageSize={20}
          categories={[{ category_id: "cat-shirts", name: "Shirts" }]}
          onPageChange={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(screen.getByText("Cotton Shirt")).toBeTruthy();
    expect(screen.getByText("Shirts")).toBeTruthy();
    expect(screen.getByText("draft")).toBeTruthy();
    expect(screen.getByRole("link", { name: /edit/i }).getAttribute("href")).toBe(
      "/seller/products/product-1/edit",
    );
  });
});
