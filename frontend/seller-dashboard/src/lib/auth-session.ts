const SELLER_SESSION_KEY = "seller-dashboard.session";

const sellerRoles = new Set([
  "seller",
  "seller_manager",
  "seller_catalog_editor",
  "seller_order_manager",
  "superadmin",
]);

export type SellerAuthSession = {
  accessToken: string;
  expiresAt?: string;
  sellerId?: string;
  userId: string;
  roles: string[];
};

function isSellerRole(role: string) {
  return sellerRoles.has(role.trim().toLowerCase());
}

export function readSellerAuthSession(): SellerAuthSession | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.sessionStorage.getItem(SELLER_SESSION_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<SellerAuthSession>;
    if (
      !parsed.accessToken ||
      !parsed.userId ||
      !Array.isArray(parsed.roles) ||
      !parsed.roles.some(isSellerRole)
    ) {
      clearSellerAuthSession();
      return null;
    }

    if (parsed.expiresAt && Date.parse(parsed.expiresAt) <= Date.now()) {
      clearSellerAuthSession();
      return null;
    }

    return {
      accessToken: parsed.accessToken,
      expiresAt: parsed.expiresAt,
      roles: parsed.roles,
      sellerId: parsed.sellerId,
      userId: parsed.userId,
    };
  } catch {
    clearSellerAuthSession();
    return null;
  }
}

export function writeSellerAuthSession(session: SellerAuthSession) {
  window.sessionStorage.setItem(SELLER_SESSION_KEY, JSON.stringify(session));
}

export function clearSellerAuthSession() {
  if (typeof window !== "undefined") {
    window.sessionStorage.removeItem(SELLER_SESSION_KEY);
  }
}

export function getSellerAccessToken() {
  return readSellerAuthSession()?.accessToken;
}
