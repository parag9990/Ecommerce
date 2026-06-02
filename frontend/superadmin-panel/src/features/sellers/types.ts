export const SELLER_STATUSES = ["draft", "pending_review", "active", "suspended", "rejected"] as const;

export type SellerStatus = (typeof SELLER_STATUSES)[number];

export const KYC_DOCUMENT_STATUSES = ["pending", "approved", "rejected"] as const;

export type KycDocumentStatus = (typeof KYC_DOCUMENT_STATUSES)[number];

export const SELLER_CATALOG_STATUSES = [
  "draft",
  "pending_review",
  "published",
  "rejected",
  "unpublished"
] as const;

export type SellerCatalogStatus = (typeof SELLER_CATALOG_STATUSES)[number];

export type SellerKycStatus = "not_started" | "pending" | "approved" | "rejected";

export type KycDocument = {
  document_id: string;
  seller_id: string;
  document_type: string;
  storage_url?: string | null;
  status: KycDocumentStatus;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  rejection_reason?: string | null;
  created_at?: string | null;
};

export type AdminSeller = {
  seller_id: string;
  user_id: string;
  store_name: string;
  display_name?: string | null;
  gst_number?: string | null;
  support_email?: string | null;
  status: SellerStatus;
  kyc_status?: SellerKycStatus;
  document_count?: number;
  pending_document_count?: number;
  product_count?: number;
  pending_catalog_count?: number;
  approved_by?: string | null;
  approved_at?: string | null;
  created_at?: string | null;
  updated_at?: string | null;
};

export type SellerCatalogItem = {
  product_id: string;
  seller_id: string;
  title: string;
  brand?: string | null;
  category_id?: string | null;
  status: SellerCatalogStatus;
  image_url?: string | null;
  updated_at?: string | null;
};

export type AdminSellerFilters = {
  q?: string;
  status?: SellerStatus | "all";
  page: number;
  page_size: number;
};

export type AdminSellerListResponse = {
  sellers: AdminSeller[];
  total?: number;
  page?: number;
  page_size?: number;
};

export type SellerKycDocumentListResponse = {
  documents: KycDocument[];
};

export type SellerCatalogResponse = {
  products: SellerCatalogItem[];
};

export type UpdateSellerStatusInput = {
  sellerId: string;
  status: Extract<SellerStatus, "active" | "suspended" | "rejected">;
  reason: string;
};

export type SuccessResponse = {
  success: boolean;
};
