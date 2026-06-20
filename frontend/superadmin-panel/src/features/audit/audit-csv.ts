import { AUDIT_EXPORT_FILENAME_PREFIX } from "./constants";
import type { AdminAuditLog } from "./types";

type AuditCsvColumn = {
  key: keyof Pick<
    AdminAuditLog,
    | "id"
    | "created_at"
    | "actor_admin_id"
    | "actor_role"
    | "action"
    | "resource_type"
    | "resource_id"
    | "request_id"
    | "ip_hash"
    | "reason"
  >;
  label: string;
};

const AUDIT_CSV_COLUMNS: AuditCsvColumn[] = [
  { key: "id", label: "id" },
  { key: "created_at", label: "created_at" },
  { key: "actor_admin_id", label: "actor_admin_id" },
  { key: "actor_role", label: "actor_role" },
  { key: "action", label: "action" },
  { key: "resource_type", label: "resource_type" },
  { key: "resource_id", label: "resource_id" },
  { key: "request_id", label: "request_id" },
  { key: "ip_hash", label: "ip_hash" },
  { key: "reason", label: "reason" }
];

export function sanitizeCsvCell(value: unknown): string {
  const text = String(value ?? "");

  return /^[=+\-@]/.test(text.trimStart()) ? `'${text}` : text;
}

function quoteCsvCell(value: unknown): string {
  const text = sanitizeCsvCell(value);
  const escaped = text.replace(/"/g, '""');

  return /[",\n\r]/.test(escaped) ? `"${escaped}"` : escaped;
}

export function auditLogsToCsv(logs: readonly AdminAuditLog[]): string {
  const header = AUDIT_CSV_COLUMNS.map((column) => quoteCsvCell(column.label)).join(",");
  const rows = logs.map((log) =>
    AUDIT_CSV_COLUMNS.map((column) => quoteCsvCell(log[column.key])).join(",")
  );

  return [header, ...rows].join("\n");
}

export function auditExportFilename(now = new Date()): string {
  return `${AUDIT_EXPORT_FILENAME_PREFIX}-${now.toISOString().slice(0, 10)}.csv`;
}

export function downloadBlob(filename: string, blob: Blob): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");

  link.href = url;
  link.download = filename;
  link.style.display = "none";
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
}

export function downloadCsv(filename: string, csv: string): void {
  downloadBlob(filename, new Blob([csv], { type: "text/csv;charset=utf-8" }));
}
