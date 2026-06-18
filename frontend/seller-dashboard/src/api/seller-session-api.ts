import { ApiError, http } from "../lib/http";

export type SellerStatus = "active" | "pending" | "suspended";

export type SellerSummary = {
  seller_id: string;
  display_name: string;
  status: SellerStatus;
  role?: string;
  roles?: string[];
  staff_role?: string;
};

export type SellerSessionResponse = {
  authenticated: boolean;
  active_seller: SellerSummary | null;
  sellers: SellerSummary[];
};

function isSellerStatus(value: unknown): value is SellerStatus {
  return value === "active" || value === "pending" || value === "suspended";
}

function normalizeSeller(value: unknown): SellerSummary | null {
  if (!value || typeof value !== "object") {
    return null;
  }

  const candidate = value as Record<string, unknown>;

  if (
    typeof candidate.seller_id !== "string" ||
    typeof candidate.display_name !== "string" ||
    !isSellerStatus(candidate.status)
  ) {
    return null;
  }

  return {
    seller_id: candidate.seller_id,
    display_name: candidate.display_name,
    status: candidate.status,
    role: typeof candidate.role === "string" ? candidate.role : undefined,
    roles: Array.isArray(candidate.roles)
      ? candidate.roles.filter((role): role is string => typeof role === "string")
      : undefined,
    staff_role:
      typeof candidate.staff_role === "string" ? candidate.staff_role : undefined,
  };
}

function normalizeSellerSession(value: unknown): SellerSessionResponse {
  if (!value || typeof value !== "object") {
    throw new Error("Seller session response was empty or invalid.");
  }

  const candidate = value as Record<string, unknown>;
  const sellers = Array.isArray(candidate.sellers)
    ? candidate.sellers.map(normalizeSeller).filter((seller): seller is SellerSummary => seller !== null)
    : [];
  const activeSeller = normalizeSeller(candidate.active_seller);

  return {
    authenticated: Boolean(candidate.authenticated),
    active_seller: activeSeller,
    sellers,
  };
}

export async function getSellerSession(): Promise<SellerSessionResponse> {
  try {
    const session = await http<SellerSessionResponse>("/api/v1/seller/session");
    return normalizeSellerSession(session);
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return {
        authenticated: false,
        active_seller: null,
        sellers: [],
      };
    }

    throw error;
  }
}
