import { fireEvent, render, screen } from "@testing-library/react";
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

  it("shows existing image URLs", () => {
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

    expect(screen.getByLabelText<HTMLInputElement>("Image URL 1").value).toBe(
      "https://cdn.example.com/one.jpg",
    );
    expect(screen.getByLabelText<HTMLInputElement>("Image URL 2").value).toBe(
      "https://cdn.example.com/two.jpg",
    );
  });

  it("updates an image URL by index", () => {
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

    fireEvent.change(screen.getByLabelText("Image URL 2"), {
      target: { value: "https://cdn.example.com/updated.jpg" },
    });

    expect(onChange).toHaveBeenCalledWith([
      "https://cdn.example.com/one.jpg",
      "https://cdn.example.com/updated.jpg",
    ]);
  });
});
