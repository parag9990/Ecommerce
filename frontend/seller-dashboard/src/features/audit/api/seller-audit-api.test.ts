import { describe, expect, it } from "vitest";

import {
  normalizeSellerAuditLog,
  normalizeSellerAuditResponse,
} from "./seller-audit-api";

const validLog = {
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
    status: "draft",
  },
  after: {
    price: 119900,
    status: "submitted",
  },
  created_at: "2026-06-01T10:30:00Z",
};

describe("seller audit api normalizers", () => {
  it("normalizes a valid audit log", () => {
    expect(normalizeSellerAuditLog(validLog)).toEqual(validLog);
  });

  it("normalizes before_json and after_json aliases", () => {
    expect(
      normalizeSellerAuditLog({
        ...validLog,
        before: undefined,
        after: undefined,
        before_json: "{\"status\":\"draft\"}",
        after_json: { status: "published" },
      }),
    ).toMatchObject({
      before: { status: "draft" },
      after: { status: "published" },
    });
  });

  it("rejects logs missing required audit fields", () => {
    expect(normalizeSellerAuditLog({ ...validLog, audit_id: 10 })).toBeNull();
    expect(normalizeSellerAuditLog({ ...validLog, action: null })).toBeNull();
  });

  it("normalizes responses and derives pagination", () => {
    expect(
      normalizeSellerAuditResponse({
        logs: [validLog, { ...validLog, audit_id: null }],
        pagination: {
          page: 2,
          page_size: 20,
          total: 45,
        },
      }),
    ).toEqual({
      logs: [validLog],
      pagination: {
        page: 2,
        page_size: 20,
        total: 45,
        has_next: true,
      },
    });
  });
});
