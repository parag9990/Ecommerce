export type SellerStaffRole =
  | "seller"
  | "seller_manager"
  | "seller_catalog_editor"
  | "seller_order_manager";

export type AssignableSellerStaffRole = Exclude<SellerStaffRole, "seller">;

export type SellerStaffStatus = "invited" | "active" | "disabled";

export type SellerPermission =
  | "dashboard:view"
  | "products:view"
  | "products:write"
  | "orders:view"
  | "orders:update_fulfillment"
  | "offers:view"
  | "offers:write"
  | "analytics:view"
  | "team:view"
  | "team:invite"
  | "team:update_role"
  | "team:disable"
  | "audit:view";

export type SellerStaffMember = {
  staff_id: string;
  seller_id: string;
  user_id: string;
  email: string;
  full_name: string | null;
  role: SellerStaffRole;
  status: SellerStaffStatus;
  invited_by: string | null;
  created_at: string;
  updated_at: string;
};

export type SellerTeamPagination = {
  page: number;
  page_size: number;
  total: number;
};

export type SellerTeamResponse = {
  members: SellerStaffMember[];
  pagination: SellerTeamPagination;
};

export type SellerTeamStatusFilter = SellerStaffStatus | "all";

export type SellerTeamFilters = {
  page: number;
  page_size: number;
  status: SellerTeamStatusFilter;
};

export type InviteStaffInput = {
  email: string;
  role: AssignableSellerStaffRole;
};

export type UpdateStaffRoleInput = {
  staff_id: string;
  role: AssignableSellerStaffRole;
};

export type DisableStaffInput = {
  staff_id: string;
  status: "disabled";
};
