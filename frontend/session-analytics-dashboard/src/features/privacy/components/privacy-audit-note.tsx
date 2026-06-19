type PrivacyAuditNoteProps = {
  disabled?: boolean;
  error?: string;
  id: string;
  label?: string;
  onChange: (value: string) => void;
  placeholder?: string;
  value: string;
};

export function PrivacyAuditNote({
  disabled,
  error,
  id,
  label = "Audit reason",
  onChange,
  placeholder = "Support ticket or compliance reason",
  value
}: PrivacyAuditNoteProps) {
  return (
    <label className="block text-sm font-medium text-zinc-700" htmlFor={id}>
      {label}
      <textarea
        className={[
          "mt-1 min-h-24 w-full rounded-md border bg-white px-3 py-2 text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10 disabled:cursor-not-allowed disabled:bg-zinc-100",
          error ? "border-red-300" : "border-zinc-300"
        ].join(" ")}
        disabled={disabled}
        id={id}
        maxLength={512}
        onChange={(event) => onChange(event.target.value)}
        placeholder={placeholder}
        value={value}
      />
      {error ? <span className="mt-1 block text-xs text-red-600">{error}</span> : null}
    </label>
  );
}
