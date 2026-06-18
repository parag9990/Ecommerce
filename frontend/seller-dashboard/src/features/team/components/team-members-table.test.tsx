import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { SellerStaffMember, SellerTeamFilters } from "../types";
import { TeamMembersTable } from "./team-members-table";

const mocks = vi.hoisted(() => ({
  can: vi.fn((permission: string) =>
    ["team:invite", "team:update_role", "team:disable"].includes(permission),
  ),
  disable: vi.fn(),
  resend: vi.fn(),
  updateRole: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("../hooks/use-seller-permissions", () => ({
  useSellerPermissions: () => ({
    roles: ["seller"],
    can: mocks.can,
  }),
}));

vi.mock("../hooks/use-seller-team", () => ({
  useDisableSellerStaff: () => ({
    isPending: false,
    mutate: mocks.disable,
  }),
  useResendSellerStaffInvite: () => ({
    isPending: false,
    mutate: mocks.resend,
  }),
  useUpdateSellerStaffRole: () => ({
    isPending: false,
    mutateAsync: mocks.updateRole,
  }),
}));

const filters: SellerTeamFilters = {
  page: 1,
  page_size: 20,
  status: "all",
};

function createMember(overrides: Partial<SellerStaffMember> = {}): SellerStaffMember {
  return {
    staff_id: "staff_123",
    seller_id: "seller_456",
    user_id: "user_789",
    email: "catalog@example.com",
    full_name: "Catalog Editor",
    role: "seller_catalog_editor",
    status: "active",
    invited_by: "owner@example.com",
    created_at: "2026-06-01T10:00:00Z",
    updated_at: "2026-06-01T10:00:00Z",
    ...overrides,
  };
}

function renderTable(members: SellerStaffMember[]) {
  return render(
    <TeamMembersTable
      members={members}
      pagination={{ page: 1, page_size: 20, total: members.length }}
      filters={filters}
      onFiltersChange={vi.fn()}
    />,
  );
}

describe("TeamMembersTable", () => {
  beforeEach(() => {
    mocks.can.mockClear();
    mocks.disable.mockClear();
    mocks.resend.mockClear();
    mocks.updateRole.mockClear();
  });

  it("updates assignable staff roles inline", async () => {
    const user = userEvent.setup();
    renderTable([createMember()]);

    await user.selectOptions(
      screen.getByLabelText("Role for catalog@example.com"),
      "seller_manager",
    );

    await waitFor(() => {
      expect(mocks.updateRole).toHaveBeenCalledWith({
        staff_id: "staff_123",
        role: "seller_manager",
      });
    });
  });

  it("resends invites for invited members", async () => {
    const user = userEvent.setup();
    renderTable([createMember({ status: "invited" })]);

    await user.click(screen.getByRole("button", { name: /resend/i }));

    expect(mocks.resend).toHaveBeenCalledWith(
      "staff_123",
      expect.objectContaining({ onError: expect.any(Function) }),
    );
  });

  it("disables active non-owner members", async () => {
    const user = userEvent.setup();
    renderTable([createMember()]);

    await user.click(screen.getByRole("button", { name: /disable/i }));

    expect(mocks.disable).toHaveBeenCalledWith(
      {
        staff_id: "staff_123",
        status: "disabled",
      },
      expect.objectContaining({ onError: expect.any(Function) }),
    );
  });

  it("protects owner and disabled member roles from inline edits", () => {
    renderTable([
      createMember({
        role: "seller",
        email: "owner@example.com",
        full_name: "Owner",
      }),
      createMember({
        staff_id: "staff_disabled",
        email: "disabled@example.com",
        status: "disabled",
      }),
    ]);

    expect(screen.getByText("Protected owner role")).toBeTruthy();
    expect(
      (screen.getByLabelText("Role for disabled@example.com") as HTMLSelectElement)
        .disabled,
    ).toBe(true);
  });
});
