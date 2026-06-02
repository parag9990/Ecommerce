import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";

import type { SellerAuditLog } from "../types";
import { ActivityEventCard } from "./activity-event-card";

const log: SellerAuditLog = {
  audit_id: "audit_123",
  seller_id: "seller_456",
  actor_user_id: "user_789",
  actor_name: "Catalog Manager",
  actor_email: "catalog@example.com",
  action: "product.updated",
  resource_type: "product",
  resource_id: "prod_101",
  resource_title: "Cotton Shirt",
  before: {
    price: 129900,
    refresh_token: "old",
  },
  after: {
    price: 119900,
    refresh_token: "new",
  },
  created_at: "2026-06-01T10:30:00Z",
};

function renderCard(auditLog: SellerAuditLog = log) {
  return render(
    <MemoryRouter>
      <ActivityEventCard log={auditLog} />
    </MemoryRouter>,
  );
}

describe("ActivityEventCard", () => {
  it("renders actor, action, linked resource, and diff summary", () => {
    renderCard();

    expect(screen.getByText("Catalog Manager")).toBeTruthy();
    expect(screen.getByText("Updated product")).toBeTruthy();
    expect(screen.getByRole("link", { name: "Cotton Shirt" })).toHaveProperty(
      "pathname",
      "/seller/products/prod_101/edit",
    );
    expect(screen.getByText("Changed fields")).toBeTruthy();
    expect(screen.getAllByText("Redacted")).toHaveLength(2);
  });

  it("renders unknown resources without a link", () => {
    renderCard({
      ...log,
      resource_type: "risk_review",
      resource_id: "risk_123",
      resource_title: undefined,
    });

    expect(screen.queryByRole("link", { name: "risk_123" })).toBeNull();
    expect(screen.getByText("risk_123")).toBeTruthy();
  });
});
