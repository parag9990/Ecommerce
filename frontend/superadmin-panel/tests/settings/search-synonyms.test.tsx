import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { SearchSynonymForm } from "../../src/features/settings/components/search-synonym-form";
import type { SearchSynonymInput } from "../../src/features/settings/types";

function FormHarness({ onSubmit }: { onSubmit: (input: SearchSynonymInput) => void }) {
  const [draft, setDraft] = useState<SearchSynonymInput>({ root: "", synonyms: [] });

  return (
    <SearchSynonymForm
      draft={draft}
      existingRoots={[]}
      canWrite
      onDraftChange={setDraft}
      onSubmit={onSubmit}
    />
  );
}

describe("SearchSynonymForm", () => {
  it("submits normalized lowercase terms without duplicates", async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    render(<FormHarness onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText(/root term/i), " Mobile ");
    await user.type(screen.getByLabelText(/synonyms/i), " Phone, smartphone, phone, mobile ");
    await user.click(screen.getByRole("button", { name: /add synonym/i }));

    expect(onSubmit).toHaveBeenCalledWith({
      root: "mobile",
      synonyms: ["phone", "smartphone"]
    });
  });
});
