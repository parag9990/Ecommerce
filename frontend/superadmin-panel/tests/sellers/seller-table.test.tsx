import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import { SellerTable } from "../../src/features/sellers/components/seller-table";
import type { AdminSeller } from "../../src/features/sellers/types";

const sellers: AdminSeller[] = [
  {
    seller_id: "seller_123",
    user_id: "user_123",
    store_name: "Acme Store",
    status: "pending_review",
    kyc_status: "pending",
    document_count: 3,
    pending_document_count: 2,
    product_count: 7,
    pending_catalog_count: 1,
    updated_at: "2026-06-01T10:00:00Z"
  }
];

function renderTable(canViewKyc: boolean) {
  render(
    <MemoryRouter>
      <SellerTable
        sellers={sellers}
        canViewKyc={canViewKyc}
        isLoading={false}
        isFetching={false}
        error={null}
        page={1}
        pageSize={20}
        totalCount={1}
        onPageChange={() => undefined}
        onRetry={() => undefined}
      />
    </MemoryRouter>
  );
}

describe("SellerTable", () => {
  it("shows seller review links and KYC counts for KYC reviewers", () => {
    renderTable(true);

    expect(screen.getByText("Acme Store")).toBeTruthy();
    expect(screen.getByText("2 pending / 3 total")).toBeTruthy();
    expect(screen.getAllByRole("link", { name: /review/i })[0]?.getAttribute("href")).toBe(
      "/admin/sellers/seller_123"
    );
  });

  it("masks KYC document counts for non-KYC roles", () => {
    renderTable(false);

    expect(screen.getByText("Masked")).toBeTruthy();
  });
});
