import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ImageUploader } from "./image-uploader";

describe("ImageUploader", () => {
  it("adds an image URL", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(<ImageUploader images={[]} onChange={onChange} />);

    await user.type(
      screen.getByPlaceholderText("https://cdn.example.com/product.jpg"),
      "https://cdn.example.com/product.jpg",
    );
    await user.click(screen.getByRole("button", { name: "Add URL" }));

    expect(onChange).toHaveBeenCalledWith(["https://cdn.example.com/product.jpg"]);
  });

  it("removes an image by index", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();

    render(
      <ImageUploader
        images={[
          "https://cdn.example.com/one.jpg",
          "https://cdn.example.com/two.jpg",
        ]}
        onChange={onChange}
      />,
    );

    await user.click(screen.getAllByRole("button", { name: "Remove image" })[0]);

    expect(onChange).toHaveBeenCalledWith(["https://cdn.example.com/two.jpg"]);
  });
});
