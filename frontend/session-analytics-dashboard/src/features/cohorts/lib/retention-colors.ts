export function getRetentionCellClass(
  rate: number,
  suppressed = false
): string {
  if (suppressed) {
    return "border-zinc-200 bg-zinc-100 text-zinc-500";
  }

  if (rate >= 60) {
    return "border-emerald-700 bg-emerald-600 text-white";
  }

  if (rate >= 40) {
    return "border-teal-600 bg-teal-500 text-white";
  }

  if (rate >= 20) {
    return "border-sky-500 bg-sky-400 text-zinc-950";
  }

  if (rate > 0) {
    return "border-amber-300 bg-amber-200 text-zinc-950";
  }

  return "border-zinc-200 bg-zinc-50 text-zinc-400";
}

export function getRetentionSummaryTone(index: number): string {
  const tones = [
    "border-sky-200 bg-sky-50 text-sky-950",
    "border-emerald-200 bg-emerald-50 text-emerald-950",
    "border-teal-200 bg-teal-50 text-teal-950",
    "border-amber-200 bg-amber-50 text-amber-950"
  ];

  return tones[index % tones.length];
}
