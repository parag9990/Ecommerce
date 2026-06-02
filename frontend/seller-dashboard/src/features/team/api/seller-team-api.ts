import { http } from "../../../lib/http";
import type {
  DisableStaffInput,
  InviteStaffInput,
  SellerStaffMember,
  SellerStaffStatus,
  SellerTeamFilters,
  SellerTeamResponse,
  UpdateStaffRoleInput,
} from "../types";
import {
  isAssignableSellerStaffRole,
  isSellerStaffRole,
} from "../utils/seller-permissions";

type RawSellerTeamResponse = {
  members?: unknown[];
  pagination?: {
    page?: number;
    page_size?: number;
    total?: number;
  };
  page?: number;
  page_size?: number;
  total?: number;
};

function isStaffStatus(value: unknown): value is SellerStaffStatus {
  return value === "invited" || value === "active" || value === "disabled";
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" ? (value as Record<string, unknown>) : null;
}

function stringOrEmpty(value: unknown) {
  return typeof value === "string" ? value : "";
}

function stringOrNull(value: unknown) {
  return typeof value === "string" && value.trim() !== "" ? value : null;
}

export function normalizeStaffMember(value: unknown): SellerStaffMember | null {
  const candidate = asRecord(value);

  if (!candidate) {
    return null;
  }

  if (
    typeof candidate.staff_id !== "string" ||
    typeof candidate.email !== "string" ||
    !isSellerStaffRole(candidate.role) ||
    !isStaffStatus(candidate.status)
  ) {
    return null;
  }

  return {
    staff_id: candidate.staff_id,
    seller_id: stringOrEmpty(candidate.seller_id),
    user_id: stringOrEmpty(candidate.user_id),
    email: candidate.email,
    full_name: stringOrNull(candidate.full_name),
    role: candidate.role,
    status: candidate.status,
    invited_by: stringOrNull(candidate.invited_by),
    created_at: stringOrEmpty(candidate.created_at),
    updated_at: stringOrEmpty(candidate.updated_at),
  };
}

function normalizePagination(
  response: RawSellerTeamResponse,
  memberCount: number,
): SellerTeamResponse["pagination"] {
  const pagination = response.pagination ?? {};

  return {
    page: Number(pagination.page ?? response.page ?? 1),
    page_size: Number(pagination.page_size ?? response.page_size ?? 20),
    total: Number(pagination.total ?? response.total ?? memberCount),
  };
}

export function normalizeSellerTeamResponse(value: unknown): SellerTeamResponse {
  const response = asRecord(value) as RawSellerTeamResponse | null;

  if (!response) {
    return {
      members: [],
      pagination: {
        page: 1,
        page_size: 20,
        total: 0,
      },
    };
  }

  const members = Array.isArray(response.members)
    ? response.members
        .map(normalizeStaffMember)
        .filter((member): member is SellerStaffMember => member !== null)
    : [];

  return {
    members,
    pagination: normalizePagination(response, members.length),
  };
}

function toQuery(params: Partial<SellerTeamFilters>) {
  const search = new URLSearchParams();

  if (params.page) {
    search.set("page", String(params.page));
  }

  if (params.page_size) {
    search.set("page_size", String(params.page_size));
  }

  if (params.status && params.status !== "all") {
    search.set("status", params.status);
  }

  const query = search.toString();

  return query ? `?${query}` : "";
}

export async function listSellerTeam(
  params: Partial<SellerTeamFilters> = {},
): Promise<SellerTeamResponse> {
  const response = await http<unknown>(`/api/v1/seller/team${toQuery(params)}`);

  return normalizeSellerTeamResponse(response);
}

export async function inviteSellerStaff(
  input: InviteStaffInput,
): Promise<SellerStaffMember> {
  const response = await http<unknown>("/api/v1/seller/team/invites", {
    method: "POST",
    body: JSON.stringify(input),
  });
  const member = normalizeStaffMember(response);

  if (!member) {
    throw new Error("Invite response did not include a valid staff member.");
  }

  return member;
}

export async function updateSellerStaffRole(
  input: UpdateStaffRoleInput,
): Promise<SellerStaffMember> {
  if (!isAssignableSellerStaffRole(input.role)) {
    throw new Error("Owner role cannot be assigned from the seller dashboard.");
  }

  const response = await http<unknown>(`/api/v1/seller/team/${input.staff_id}/role`, {
    method: "PATCH",
    body: JSON.stringify({ role: input.role }),
  });
  const member = normalizeStaffMember(response);

  if (!member) {
    throw new Error("Role update response did not include a valid staff member.");
  }

  return member;
}

export async function disableSellerStaff(
  input: DisableStaffInput,
): Promise<SellerStaffMember> {
  const response = await http<unknown>(`/api/v1/seller/team/${input.staff_id}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status: input.status }),
  });
  const member = normalizeStaffMember(response);

  if (!member) {
    throw new Error("Status update response did not include a valid staff member.");
  }

  return member;
}

export async function resendSellerStaffInvite(staffId: string): Promise<SellerStaffMember> {
  const response = await http<unknown>(`/api/v1/seller/team/${staffId}/resend-invite`, {
    method: "POST",
  });
  const member = normalizeStaffMember(response);

  if (!member) {
    throw new Error("Resend invite response did not include a valid staff member.");
  }

  return member;
}
