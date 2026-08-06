import type {
  Category,
  Product,
  ProductInput,
  ProductStatus,
  ProductVariantInput,
} from "../types";
import { PRODUCT_STATUSES } from "../types";
import type { ProductFormValues } from "./product-validation";
import { createEmptyVariant } from "./product-validation";

function optionalTrim(value?: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function cleanAttributes(attributes?: Record<string, string>) {
  return Object.fromEntries(
    Object.entries(attributes ?? {})
      .map(([key, value]) => [key.trim(), value.trim()] as const)
      .filter(([key, value]) => key.length > 0 && value.length > 0),
  );
}

function normalizeString(value: unknown) {
  return typeof value === "string" ? value : "";
}

function normalizeOptionalString(value: unknown) {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function normalizeOptionalTrimmedString(value: unknown) {
  if (typeof value !== "string") {
    return undefined;
  }

  const trimmed = value.trim();
  return trimmed ? trimmed : undefined;
}

function normalizeNumber(value: unknown) {
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : 0;
}

function normalizeAttributes(value: unknown): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  return Object.fromEntries(
    Object.entries(value as Record<string, unknown>)
      .map(([key, entryValue]) => [key, String(entryValue ?? "")] as const)
      .filter(([key]) => key.length > 0),
  );
}

type NormalizedImage = {
  isPrimary: boolean;
  position: number;
  status: string;
  url: string;
};

function normalizeImage(value: unknown, index: number): NormalizedImage | undefined {
  if (typeof value === "string") {
    const url = normalizeOptionalTrimmedString(value);

    return url
      ? {
          isPrimary: index === 0,
          position: index,
          status: "active",
          url,
        }
      : undefined;
  }

  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return undefined;
  }

  const candidate = value as Record<string, unknown>;
  const url = normalizeOptionalTrimmedString(candidate.url);
  const position = Number(candidate.position);

  if (!url) {
    return undefined;
  }

  return {
    isPrimary: candidate.is_primary === true || candidate.isPrimary === true,
    position: Number.isFinite(position) ? position : index,
    status: normalizeOptionalTrimmedString(candidate.status) ?? "active",
    url,
  };
}

function normalizeImages(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }

  return value
    .map(normalizeImage)
    .filter((image): image is NormalizedImage => Boolean(image?.url))
    .filter((image) => image.status === "active" || image.status === "")
    .sort((left, right) => {
      if (left.isPrimary !== right.isPrimary) {
        return left.isPrimary ? -1 : 1;
      }

      return left.position - right.position;
    })
    .map((image) => image.url);
}

export function isProductStatus(value: unknown): value is ProductStatus {
  return PRODUCT_STATUSES.some((status) => status === value);
}

function normalizeProductStatus(value: unknown): ProductStatus {
  return isProductStatus(value) ? value : "draft";
}

function normalizeVariant(value: unknown): ProductVariantInput {
  const candidate = value && typeof value === "object" ? (value as Record<string, unknown>) : {};
  const price =
    candidate.price && typeof candidate.price === "object"
      ? (candidate.price as Record<string, unknown>)
      : {};

  return {
    sku: normalizeString(candidate.sku),
    attributes: normalizeAttributes(candidate.attributes),
    price: {
      amount: normalizeNumber(price.amount),
      currency: "INR",
    },
    stock_quantity: normalizeNumber(candidate.stock_quantity),
  };
}

export function toProductInput(values: ProductFormValues): ProductInput {
  return {
    title: values.title.trim(),
    description: optionalTrim(values.description),
    brand: optionalTrim(values.brand),
    category_id: values.category_id,
    attributes: cleanAttributes(values.attributes),
    images: values.images.map((image) => image.trim()).filter(Boolean),
    variants: values.variants.map((variant) => ({
      sku: variant.sku.trim(),
      attributes: cleanAttributes(variant.attributes),
      price: {
        amount: Number(variant.price.amount),
        currency: "INR",
      },
      stock_quantity: Number(variant.stock_quantity),
    })),
  };
}

export function toProductFormValues(product?: Product): ProductFormValues {
  return {
    title: product?.title ?? "",
    description: product?.description ?? "",
    brand: product?.brand ?? "",
    category_id: product?.category_id ?? "",
    attributes: product?.attributes ?? {},
    images: product?.images ?? [],
    variants: product?.variants.length ? product.variants : [createEmptyVariant()],
  };
}

export function normalizeProduct(value: unknown): Product {
  if (!value || typeof value !== "object") {
    throw new Error("Product response was empty or invalid.");
  }

  const candidate = value as Record<string, unknown>;
  const variants = Array.isArray(candidate.variants)
    ? candidate.variants.map(normalizeVariant)
    : [];

  return {
    product_id: normalizeString(candidate.product_id),
    seller_id: normalizeString(candidate.seller_id),
    title: normalizeString(candidate.title),
    description: normalizeOptionalString(candidate.description),
    brand: normalizeOptionalString(candidate.brand),
    category_id: normalizeString(candidate.category_id),
    attributes: normalizeAttributes(candidate.attributes),
    images: normalizeImages(candidate.images),
    variants,
    status: normalizeProductStatus(candidate.status),
    created_at: normalizeOptionalString(candidate.created_at),
    updated_at: normalizeOptionalString(candidate.updated_at),
  };
}

export function normalizeCategory(value: unknown): Category | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const candidate = value as Record<string, unknown>;

  if (typeof candidate.category_id !== "string" || typeof candidate.name !== "string") {
    return null;
  }

  return {
    category_id: candidate.category_id,
    name: candidate.name,
    parent_id: normalizeOptionalString(candidate.parent_id),
  };
}
