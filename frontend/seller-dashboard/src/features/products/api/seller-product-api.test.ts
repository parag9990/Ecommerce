import { afterEach, describe, expect, it, vi } from "vitest";

import { clearSellerAuthSession } from "../../../lib/auth-session";
import { getProduct, listSellerProducts } from "./seller-product-api";

const draftProduct = {
  product_id: "prod_1",
  seller_id: "seller_1",
  title: "Draft shirt",
  description: "Draft product",
  brand: "Codex",
  category_id: "cat_mens_shirts",
  attributes: {},
  images: [],
  variants: [
    {
      sku: "DRAFT-SHIRT-M",
      attributes: {},
      price: { amount: 999, currency: "INR" },
      stock_quantity: 5,
    },
  ],
  status: "draft",
};

function mockJSONResponse(body: unknown) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "content-type": "application/json" },
  });
}

describe("seller product api", () => {
  afterEach(() => {
    clearSellerAuthSession();
    vi.restoreAllMocks();
  });

  it("lists products through the seller-scoped product route", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      mockJSONResponse({ products: [draftProduct], total: 1 }),
    );

    await expect(
      listSellerProducts({
        seller_id: "seller_1",
        status: "all",
        page: 1,
        page_size: 20,
      }),
    ).resolves.toMatchObject({ total: 1, products: [{ product_id: "prod_1" }] });

    const requestUrl = String(fetchMock.mock.calls[0]?.[0]);
    expect(requestUrl).toBe("http://localhost:8080/api/v1/seller/products?page=1&page_size=20");
  });

  it("loads editable product details through the seller-scoped route", async () => {
    const fetchMock = vi.spyOn(globalThis, "fetch").mockResolvedValue(
      mockJSONResponse(draftProduct),
    );

    await expect(getProduct("prod_1")).resolves.toMatchObject({
      product_id: "prod_1",
      status: "draft",
    });

    expect(String(fetchMock.mock.calls[0]?.[0])).toBe(
      "http://localhost:8080/api/v1/seller/products/prod_1",
    );
  });
});
