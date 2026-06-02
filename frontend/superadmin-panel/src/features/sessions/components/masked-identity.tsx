export function maskIdentifier(value?: string | null): string {
  const trimmed = value?.trim();

  if (!trimmed) {
    return "Not linked";
  }

  if (trimmed.length <= 4) {
    return "****";
  }

  if (trimmed.length <= 8) {
    return `${trimmed.slice(0, 2)}...${trimmed.slice(-2)}`;
  }

  return `${trimmed.slice(0, 4)}...${trimmed.slice(-4)}`;
}

export function MaskedIdentity({
  value,
  emptyLabel = "Not linked",
  label = "Identifier"
}: {
  value?: string | null;
  emptyLabel?: string;
  label?: string;
}) {
  const maskedValue = value ? maskIdentifier(value) : emptyLabel;

  return (
    <code
      aria-label={`${label}: ${maskedValue}`}
      className="inline-flex max-w-full rounded bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-700"
    >
      <span className="truncate">{maskedValue}</span>
    </code>
  );
}
