import { Search } from "lucide-react";
import { type FormEvent, useEffect, useState } from "react";

type JourneySessionSearchProps = {
  initialValue: string;
  onSubmit: (sessionId: string) => void;
};

export function JourneySessionSearch({
  initialValue,
  onSubmit
}: JourneySessionSearchProps) {
  const [value, setValue] = useState(initialValue);

  useEffect(() => {
    setValue(initialValue);
  }, [initialValue]);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextValue = value.trim();

    if (nextValue) {
      onSubmit(nextValue);
    }
  }

  return (
    <form
      className="grid gap-2 sm:grid-cols-[minmax(220px,320px)_auto]"
      onSubmit={handleSubmit}
    >
      <label className="block" htmlFor="journey-session-id">
        <span className="sr-only">Session id</span>
        <input
          className="h-10 w-full rounded-md border border-zinc-300 bg-white px-3 font-mono text-sm outline-none transition-colors focus:border-zinc-950 focus:ring-2 focus:ring-zinc-950/10"
          id="journey-session-id"
          maxLength={160}
          onChange={(event) => setValue(event.target.value)}
          placeholder="sess_123"
          value={value}
        />
      </label>
      <button
        className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-zinc-950 px-4 text-sm font-medium text-white transition-colors hover:bg-zinc-800 disabled:cursor-not-allowed disabled:bg-zinc-400"
        disabled={!value.trim()}
        type="submit"
      >
        <Search className="h-4 w-4" aria-hidden="true" />
        Open
      </button>
    </form>
  );
}
