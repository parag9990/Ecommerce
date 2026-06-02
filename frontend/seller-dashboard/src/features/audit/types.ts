export type AuditResourceType =
  | "product"
  | "variant"
  | "order"
  | "coupon"
  | "campaign"
  | "team"
  | "settings";

export type AuditAction =
  | "product.created"
  | "product.updated"
  | "product.submitted"
  | "product.published"
  | "product.unpublished"
  | "variant.created"
  | "variant.updated"
  | "variant.stock_updated"
  | "order.status_updated"
  | "order.shipment_updated"
  | "order.refund_viewed"
  | "coupon.created"
  | "coupon.updated"
  | "coupon.paused"
  | "coupon.activated"
  | "campaign.created"
  | "campaign.updated"
  | "campaign.paused"
  | "team.member_invited"
  | "team.role_updated"
  | "team.member_disabled"
  | "settings.updated"
  | (string & {});

export type AuditJson = Record<string, unknown> | null;

export type SellerAuditLog = {
  audit_id: string;
  seller_id: string;
  actor_user_id: string;
  actor_name?: string;
  actor_email?: string;
  action: AuditAction;
  resource_type: AuditResourceType | (string & {});
  resource_id: string;
  resource_title?: string;
  before: AuditJson;
  after: AuditJson;
  created_at: string;
};

export type AuditFilters = {
  page_size?: number;
  actor_id?: string;
  action?: string;
  resource_type?: AuditResourceType | string;
  resource_id?: string;
  from?: string;
  to?: string;
};

export type AuditListParams = AuditFilters & {
  page?: number;
};

export type AuditPagination = {
  page: number;
  page_size: number;
  total?: number;
  has_next: boolean;
};

export type SellerAuditLogResponse = {
  logs: SellerAuditLog[];
  pagination: AuditPagination;
};

export const DEFAULT_AUDIT_PAGE_SIZE = 20;
