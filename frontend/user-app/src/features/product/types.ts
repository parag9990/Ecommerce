export type Money = {
  amount?: number | undefined;
  currency?: string | undefined;
};

export type ProductVariant = {
  attributes?: Record<string, unknown> | undefined;
  available_quantity?: number | undefined;
  price?: Money | undefined;
  sku: string;
  stock_quantity?: number | undefined;
  variant_id?: string | undefined;
};

export type Product = {
  brand?: string | undefined;
  category_id?: string | undefined;
  category_ids?: string[] | undefined;
  description?: string | undefined;
  images?: string[] | undefined;
  product_id: string;
  rating?: number | undefined;
  seller_id?: string | undefined;
  status?: string | undefined;
  title: string;
  variants?: ProductVariant[] | undefined;
};

export type Category = {
  category_id: string;
  name: string;
  parent_id?: string | undefined;
};

export type ProductListResponse = {
  products: Product[];
  total: number;
};

export type CategoryListResponse = {
  categories: Category[];
};

export type SearchResponse = {
  facets?: Record<string, unknown> | undefined;
  products: Product[];
  total: number;
};

export type AutocompleteResponse = {
  suggestions: string[];
};

export const productSortOptions = [
  'popular',
  'newest',
  'price_asc',
  'price_desc',
  'rating_desc',
] as const;

export type SortOption = (typeof productSortOptions)[number];

export type ProductFilters = {
  brand?: string | undefined;
  categoryId?: string | undefined;
  inStock?: boolean | undefined;
  maxPrice?: string | undefined;
  minPrice?: string | undefined;
  minRating?: string | undefined;
  page: number;
  pageSize: number;
  q?: string | undefined;
  sellerId?: string | undefined;
  sort?: SortOption | undefined;
};
