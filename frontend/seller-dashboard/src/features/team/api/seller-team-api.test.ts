import { describe, expect, it } from "vitest";

import { normalizeSellerTeamResponse, normalizeStaffMember } from "./seller-team-api";

const validMember = {
  staff_id: "staff_123",
  seller_id: "seller_456",
  user_id: "user_789",
  email: "catalog@example.com",
  full_name: "Catalog Editor",
  role: "seller_catalog_editor",
  status: "active",
  invited_by: "user_owner",
  created_at: "2026-06-01T10:00:00Z",
  updated_at: "2026-06-01T10:00:00Z",
};

describe("seller team api normalizers", () => {
  it("normalizes a valid staff member", () => {
    expect(normalizeStaffMember(validMember)).toEqual(validMember);
  });

  it("rejects staff members with unknown role or status", () => {
    expect(normalizeStaffMember({ ...validMember, role: "admin" })).toBeNull();
    expect(normalizeStaffMember({ ...validMember, status: "archived" })).toBeNull();
  });

  it("normalizes team responses and skips invalid members", () => {
    expect(
      normalizeSellerTeamResponse({
        members: [validMember, { ...validMember, staff_id: 10 }],
        pagination: {
          page: 2,
          page_size: 10,
          total: 11,
        },
      }),
    ).toEqual({
      members: [validMember],
      pagination: {
        page: 2,
        page_size: 10,
        total: 11,
      },
    });
  });
});
