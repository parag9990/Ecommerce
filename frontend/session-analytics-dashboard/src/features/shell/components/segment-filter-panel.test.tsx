import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { defaultSegmentFilters } from "../types";
import { SegmentFilterPanel } from "./segment-filter-panel";

describe("SegmentFilterPanel", () => {
  it("emits typed filter changes", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <SegmentFilterPanel value={defaultSegmentFilters} onChange={onChange} />
    );

    await user.selectOptions(screen.getByLabelText("Device"), "mobile");

    expect(onChange).toHaveBeenCalledWith({
      ...defaultSegmentFilters,
      deviceType: "mobile"
    });
  });
});
