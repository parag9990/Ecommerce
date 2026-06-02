export const PRODUCT_STATUSES = [
  "draft",
  "submitted",
  "approved",
  "rejected",
  "published",
  "unpublished",
] as const;

export type ProductStatus = (typeof PRODUCT_STATUSES)[number];

export type ProductStatusFilter = ProductStatus | "all";

export type Money = {
  amount: number;
  currency: "INR";
};

export type ProductVariantInput = {
  sku: string;
  attributes: Record<string, string>;
  price: Money;
  stock_quantity: number;
};

export type ProductInput = {
  title: string;
  description?: string;
  brand?: string;
  category_id: string;
  attributes: Record<string, string>;
  images: string[];
  variants: ProductVariantInput[];
};

export type Product = ProductInput & {
  product_id: string;
  seller_id: string;
  status: ProductStatus;
  created_at?: string;
  updated_at?: string;
};

export type ProductListRequest = {
  seller_id: string;
  category_id?: string;
  status?: ProductStatusFilter;
  page: number;
  page_size: number;
};

export type ProductListResponse = {
  products: Product[];
  total: number;
};

export type Category = {
  category_id: string;
  name: string;
  parent_id?: string;
};
