import { Timeline } from "../../../components/ui/timeline";
import { formatDateTime } from "../../../lib/format";
import type { PaymentAttempt } from "../types";

export function PaymentAttemptsTimeline({ attempts }: { attempts: PaymentAttempt[] }) {
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-4">
      <h2 className="text-base font-semibold text-slate-950">Attempt Timeline</h2>
      <p className="mt-1 text-sm text-slate-600">
        Provider attempts and retry outcomes. Webhooks remain the final source of truth.
      </p>
      <div className="mt-4">
        <Timeline
          emptyLabel="No payment attempts were returned for this payment."
          items={attempts.map((attempt, index) => ({
            id: `${attempt.attempt_id}-${index}`,
            title: attempt.status.replace(/_/g, " "),
            description: (
              <>
                {attempt.provider_reference ? <span>Provider ref {attempt.provider_reference}</span> : null}
                {attempt.error_code ? <span> Error {attempt.error_code}</span> : null}
                {attempt.error_message ? <span> - {attempt.error_message}</span> : null}
              </>
            ),
            timestamp: formatDateTime(attempt.created_at)
          }))}
        />
      </div>
    </section>
  );
}
