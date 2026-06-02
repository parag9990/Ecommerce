import { apiFetch } from "../../../lib/http";
import type {
  AdminSeller,
  AdminSellerFilters,
  AdminSellerListResponse,
  KycDocument,
  SellerCatalogItem,
  SellerCatalogResponse,
  SellerKycDocumentListResponse,
  SellerStatus,
  SuccessResponse,
  UpdateSellerStatusInput
} from "../types";

const SELLERS_PATH = "/api/v1/admin/sellers";
const DEFAULT_SELLER_STATUS: SellerStatus = "draft";

function buildQueryString(values: Record<string, string | number | undefined>): string {
  const params = new URLSearchParams();

  Object.entries(values).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      params.set(key, String(value));
    }
  });

  return params.toString();
}

function normalizeSeller(seller: AdminSeller): AdminSeller {
  return {
    ...seller,
    status: seller.status ?? DEFAULT_SELLER_STATUS,
    store_name: seller.store_name?.trim() || "Unnamed store"
  };
}

function normalizeKycDocument(document: KycDocument): KycDocument {
  return {
    ...document,
    status: document.status ?? "pending",
    document_type: document.document_type?.trim() || "KYC document"
  };
}

function normalizeCatalogItem(product: SellerCatalogItem): SellerCatalogItem {
  return {
    ...product,
    status: product.status ?? "draft",
    title: product.title?.trim() || "Untitled product"
  };
}

export function toAdminSellersQueryString(filters: AdminSellerFilters): string {
  return buildQueryString({
    q: filters.q?.trim(),
    status: filters.status && filters.status !== "all" ? filters.status : undefined,
    page: filters.page,
    page_size: filters.page_size
  });
}

export async function listAdminSellers(filters: AdminSellerFilters): Promise<AdminSellerListResponse> {
  const query = toAdminSellersQueryString(filters);
  const response = await apiFetch<AdminSellerListResponse>(`${SELLERS_PATH}?${query}`);

  return {
    ...response,
    sellers: (response.sellers ?? []).map(normalizeSeller),
    page: response.page ?? filters.page,
    page_size: response.page_size ?? filters.page_size
  };
}

export async function getAdminSeller(sellerId: string): Promise<AdminSeller | null> {
  const response = await listAdminSellers({
    q: sellerId,
    status: "all",
    page: 1,
    page_size: 1
  });

  return response.sellers.find((seller) => seller.seller_id === sellerId) ?? response.sellers[0] ?? null;
}

export async function updateSellerStatus(input: UpdateSellerStatusInput): Promise<SuccessResponse> {
  return apiFetch<SuccessResponse>(`${SELLERS_PATH}/${encodeURIComponent(input.sellerId)}/status`, {
    method: "PATCH",
    body: JSON.stringify({
      status: input.status,
      reason: input.reason
    })
  });
}

export async function listSellerKycDocuments(sellerId: string): Promise<SellerKycDocumentListResponse> {
  const response = await apiFetch<SellerKycDocumentListResponse>(
    `${SELLERS_PATH}/${encodeURIComponent(sellerId)}/kyc-documents`
  );

  return {
    ...response,
    documents: (response.documents ?? []).map(normalizeKycDocument)
  };
}

export async function listSellerCatalog(sellerId: string): Promise<SellerCatalogResponse> {
  const response = await apiFetch<SellerCatalogResponse>(
    `${SELLERS_PATH}/${encodeURIComponent(sellerId)}/catalog`
  );

  return {
    ...response,
    products: (response.products ?? []).map(normalizeCatalogItem)
  };
}
