import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useCategories } from "../hooks/use-categories";
import type { Category } from "../types";
import { CategorySelector } from "./category-selector";

vi.mock("../hooks/use-categories", () => ({
  useCategories: vi.fn(),
}));

type MockCategoriesQuery = {
  data?: Category[];
  isLoading: boolean;
  isError: boolean;
};

const mockedUseCategories = vi.mocked(useCategories);

function mockCategoriesQuery(query: MockCategoriesQuery) {
  mockedUseCategories.mockReturnValue(query as ReturnType<typeof useCategories>);
}

describe("CategorySelector", () => {
  beforeEach(() => {
    mockCategoriesQuery({
      data: [
        { category_id: "cat_mens_shirts", name: "Men's Shirts" },
        { category_id: "cat_home_decor", name: "Home Decor" },
      ],
      isLoading: false,
      isError: false,
    });
  });

  it("renders active categories and emits the selected id", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<CategorySelector value="" onChange={onChange} />);

    expect((screen.getByRole("option", { name: "Men's Shirts" }) as HTMLOptionElement).value).toBe(
      "cat_mens_shirts",
    );
    expect((screen.getByRole("option", { name: "Home Decor" }) as HTMLOptionElement).value).toBe(
      "cat_home_decor",
    );

    await user.selectOptions(screen.getByLabelText("Category"), "cat_home_decor");

    expect(onChange).toHaveBeenCalledWith("cat_home_decor");
  });

  it("shows an explicit empty state when no active categories are returned", () => {
    mockCategoriesQuery({ data: [], isLoading: false, isError: false });

    render(<CategorySelector value="" onChange={vi.fn()} />);

    expect((screen.getByRole("combobox") as HTMLSelectElement).disabled).toBe(true);
    expect(screen.getByText("No active categories are available.").textContent).toBe(
      "No active categories are available.",
    );
  });

  it("keeps the control disabled while categories are loading", () => {
    mockCategoriesQuery({ data: undefined, isLoading: true, isError: false });

    render(<CategorySelector value="" onChange={vi.fn()} />);

    expect((screen.getByRole("combobox") as HTMLSelectElement).disabled).toBe(true);
    expect(screen.getAllByText("Loading categories...").length).toBeGreaterThan(0);
  });
});
