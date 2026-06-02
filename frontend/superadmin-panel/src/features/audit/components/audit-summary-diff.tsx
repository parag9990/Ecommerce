import type { AuditSummary } from "../types";

const SENSITIVE_KEY_PATTERN =
  /(password|passcode|otp|token|secret|credential|authorization|cookie|card|cvv|private_key)/i;

function sanitizeSummaryValue(value: unknown, parentKey = "", depth = 0): unknown {
  if (SENSITIVE_KEY_PATTERN.test(parentKey)) {
    return "[masked]";
  }

  if (depth > 8) {
    return "[max depth]";
  }

  if (Array.isArray(value)) {
    return value.map((item) => sanitizeSummaryValue(item, parentKey, depth + 1));
  }

  if (typeof value === "object" && value !== null) {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([key, nestedValue]) => [
        key,
        sanitizeSummaryValue(nestedValue, key, depth + 1)
      ])
    );
  }

  return value;
}

function prettySummary(value: AuditSummary | null): string {
  if (!value || Object.keys(value).length === 0) {
    return "No summary available";
  }

  return JSON.stringify(sanitizeSummaryValue(value), null, 2);
}

export function AuditSummaryDiff({
  before,
  after
}: {
  before: AuditSummary | null;
  after: AuditSummary | null;
}) {
  return (
    <section className="grid gap-4">
      <div>
        <h3 className="text-sm font-semibold text-slate-950">Before Summary</h3>
        <pre className="mt-2 max-h-72 overflow-auto rounded-lg bg-slate-950 p-3 text-xs leading-5 text-slate-100">
          {prettySummary(before)}
        </pre>
      </div>

      <div>
        <h3 className="text-sm font-semibold text-slate-950">After Summary</h3>
        <pre className="mt-2 max-h-72 overflow-auto rounded-lg bg-slate-950 p-3 text-xs leading-5 text-slate-100">
          {prettySummary(after)}
        </pre>
      </div>
    </section>
  );
}
