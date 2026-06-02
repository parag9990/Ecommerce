import type { AuditResourceType } from "../types";

export const auditActionLabels: Record<string, string> = {
  "product.created": "Created product",
  "product.updated": "Updated product",
  "product.submitted": "Submitted product for review",
  "product.published": "Published product",
  "product.unpublished": "Unpublished product",
  "variant.created": "Created variant",
  "variant.updated": "Updated variant",
  "variant.stock_updated": "Updated stock",
  "order.status_updated": "Updated order status",
  "order.shipment_updated": "Updated shipment",
  "order.refund_viewed": "Viewed refund",
  "coupon.created": "Created coupon",
  "coupon.updated": "Updated coupon",
  "coupon.paused": "Paused coupon",
  "coupon.activated": "Activated coupon",
  "campaign.created": "Created campaign",
  "campaign.updated": "Updated campaign",
  "campaign.paused": "Paused campaign",
  "team.member_invited": "Invited team member",
  "team.role_updated": "Updated team role",
  "team.member_disabled": "Disabled team member",
  "settings.updated": "Updated store settings",
};

export const auditResourceLabels: Record<AuditResourceType, string> = {
  product: "Product",
  variant: "Variant",
  order: "Order",
  coupon: "Coupon",
  campaign: "Campaign",
  team: "Team",
  settings: "Settings",
};

function titleCase(value: string) {
  return value
    .split(/[\s._-]+/)
    .filter(Boolean)
    .map((word) => `${word.slice(0, 1).toUpperCase()}${word.slice(1)}`)
    .join(" ");
}

export function getActionLabel(action: string) {
  return auditActionLabels[action] ?? titleCase(action);
}

export function getResourceLabel(resourceType: string) {
  return auditResourceLabels[resourceType as AuditResourceType] ?? titleCase(resourceType);
}
