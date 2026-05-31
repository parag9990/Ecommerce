import type { FunnelReport } from "../lib/funnel-math";
import { FunnelStepCard } from "./funnel-step-card";

type FunnelStepListProps = {
  report: FunnelReport;
};

export function FunnelStepList({ report }: FunnelStepListProps) {
  return (
    <section
      aria-label="Funnel steps"
      className="grid gap-4 md:grid-cols-2 xl:grid-cols-4"
    >
      {report.steps.map((step, index) => (
        <FunnelStepCard
          index={index}
          key={step.key}
          minSegmentSize={report.minSegmentSize}
          step={step}
        />
      ))}
    </section>
  );
}
