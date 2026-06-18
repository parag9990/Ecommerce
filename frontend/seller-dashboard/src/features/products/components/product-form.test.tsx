import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ProductForm } from "./product-form";

vi.mock("../hooks/use-categories", () => ({
  useCategories: () => ({
    data: [{ category_id: "cat-shirts", name: "Shirts" }],
    isLoading: false,
    isError: false,
  }),
}));

describe("ProductForm", () => {
  it("submits a valid product input", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);

    render(
      <ProductForm
        mode="create"
        saving={false}
        onSubmit={onSubmit}
      />,
    );

    await user.type(screen.getByLabelText("Title"), "Cotton Shirt");
    await user.type(screen.getByLabelText("Brand"), "North");
    await user.selectOptions(screen.getByLabelText("Category"), "cat-shirts");
    await user.type(screen.getByLabelText("Description"), "Soft cotton shirt");
    await user.type(screen.getByLabelText("SKU"), "SHIRT-BLK-M");
    await user.clear(screen.getByLabelText("Price"));
    await user.type(screen.getByLabelText("Price"), "999");
    await user.clear(screen.getByLabelText("Stock"));
    await user.type(screen.getByLabelText("Stock"), "12");
    await user.click(screen.getByRole("button", { name: "Save draft" }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        title: "Cotton Shirt",
        description: "Soft cotton shirt",
        brand: "North",
        category_id: "cat-shirts",
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
      });
    });
  });
});
