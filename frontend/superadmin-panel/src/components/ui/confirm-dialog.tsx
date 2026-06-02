import type { ReactNode } from "react";

import { cn } from "../../lib/classnames";

export function ConfirmDialog({
  open,
  title,
  description,
  children,
  footer,
  tone = "neutral",
  onClose
}: {
  open: boolean;
  title: string;
  description?: string;
  children?: ReactNode;
  footer: ReactNode;
  tone?: "neutral" | "danger";
  onClose: () => void;
}) {
  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4">
      <button
        type="button"
        className="absolute inset-0 cursor-default"
        aria-label="Close dialog"
        onClick={onClose}
      />
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        className="relative w-full max-w-md rounded-lg border border-slate-200 bg-white p-4 shadow-xl"
      >
        <h2
          id="confirm-dialog-title"
          className={cn("text-base font-semibold", tone === "danger" ? "text-red-800" : "text-slate-950")}
        >
          {title}
        </h2>
        {description ? <p className="mt-1 text-sm text-slate-600">{description}</p> : null}
        {children ? <div className="mt-4">{children}</div> : null}
        <div className="mt-4 flex justify-end gap-2">{footer}</div>
      </section>
    </div>
  );
}
