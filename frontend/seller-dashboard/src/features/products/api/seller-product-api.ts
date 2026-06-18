import { http } from "../../../lib/http";
import type {
  Category,
  Product,
  ProductInput,
  ProductListRequest,
  ProductListResponse,
} from "../types";
import { normalizeCategory, normalizeProduct } from "../utils/product-mappers";

type CategoryListResponse = {
  categories?: unknown[];
};

type RawProductListResponse = {
  products?: unknown[];
  total?: number;
};

function toQuery(params: Record<string, string | number | undefined>) {
  const search = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });

  return search.toString();
}

export async function listSellerProducts(
  params: ProductListRequest,
): Promise<ProductListResponse> {
  const query = toQuery({
    seller_id: params.seller_id,
    category_id: params.category_id,
    status: params.status === "all" ? undefined : params.status,
    page: params.page,
    page_size: params.page_size,
  });
  const response = await http<RawProductListResponse>(`/api/v1/products?${query}`);
  const products = Array.isArray(response.products)
    ? response.products.map(normalizeProduct)
    : [];

  return {
    products,
    total: Number(response.total ?? products.length),
  };
}

export async function getProduct(productId: string): Promise<Product> {
  const product = await http<unknown>(`/api/v1/products/${productId}`);
  return normalizeProduct(product);
}

export async function listCategories(): Promise<Category[]> {
  const response = await http<CategoryListResponse>("/api/v1/categories");
  const categories = Array.isArray(response.categories) ? response.categories : [];

  return categories.map(normalizeCategory).filter((category): category is Category => Boolean(category));
}

export async function createProduct(input: ProductInput): Promise<Product> {
  const product = await http<unknown>("/api/v1/seller/products", {
    method: "POST",
    body: JSON.stringify(input),
  });

  return normalizeProduct(product);
}

export async function updateProduct(
  productId: string,
  input: ProductInput,
): Promise<Product> {
  const product = await http<unknown>(`/api/v1/seller/products/${productId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });

  return normalizeProduct(product);
}

export async function publishProduct(productId: string): Promise<Product> {
  const product = await http<unknown>(`/api/v1/seller/products/${productId}/publish`, {
    method: "POST",
  });

  return normalizeProduct(product);
}
