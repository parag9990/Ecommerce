import { ImagePlus, LinkIcon, Trash2 } from "lucide-react";
import { useState } from "react";

import { IconButton } from "../../../components/ui/icon-button";

type ImageUploaderProps = {
  images: string[];
  error?: string;
  onChange: (images: string[]) => void;
  onUpload?: (file: File) => Promise<string>;
};

export function ImageUploader({ images, error, onChange, onUpload }: ImageUploaderProps) {
  const [imageUrl, setImageUrl] = useState("");
  const [uploading, setUploading] = useState(false);

  function addImageUrl() {
    const nextImage = imageUrl.trim();

    if (!nextImage || images.includes(nextImage)) {
      return;
    }

    onChange([...images, nextImage]);
    setImageUrl("");
  }

  async function handleFileChange(file?: File) {
    if (!file || !onUpload) {
      return;
    }

    setUploading(true);

    try {
      const uploadedUrl = await onUpload(file);
      onChange([...images, uploadedUrl]);
    } finally {
      setUploading(false);
    }
  }

  function removeImage(index: number) {
    onChange(images.filter((_, imageIndex) => imageIndex !== index));
  }

  function updateImage(index: number, nextUrl: string) {
    onChange(images.map((image, imageIndex) => (imageIndex === index ? nextUrl : image)));
  }

  return (
    <section className="space-y-3 border-t border-slate-100 pt-4">
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-slate-950">Images</h2>
        <span className="text-xs font-medium text-slate-500">{images.length} added</span>
      </div>

      <div className="flex flex-wrap gap-2">
        <div className="relative min-w-64 flex-1">
          <LinkIcon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <input
            value={imageUrl}
            onChange={(event) => setImageUrl(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                event.preventDefault();
                addImageUrl();
              }
            }}
            className="h-10 w-full rounded-md border border-slate-300 pl-9 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            placeholder="https://cdn.example.com/product.jpg"
          />
        </div>
        <button
          type="button"
          onClick={addImageUrl}
          className="inline-flex h-10 items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600"
        >
          <LinkIcon className="h-4 w-4" aria-hidden="true" />
          Add URL
        </button>

        {onUpload ? (
          <label className="inline-flex h-10 cursor-pointer items-center gap-2 rounded-md border border-slate-200 bg-white px-3 text-sm font-medium text-slate-700 transition hover:bg-slate-50 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-blue-600">
            <ImagePlus className="h-4 w-4" aria-hidden="true" />
            {uploading ? "Uploading..." : "Upload"}
            <input
              type="file"
              accept="image/*"
              className="sr-only"
              onChange={(event) => handleFileChange(event.target.files?.[0])}
            />
          </label>
        ) : null}
      </div>

      {error ? <p className="text-xs text-rose-600">{error}</p> : null}

      {images.length ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {images.map((src, index) => (
            <div key={index} className="overflow-hidden rounded-md border border-slate-200 bg-white">
              <div className="group relative bg-slate-50">
                <img src={src} alt="" className="aspect-square w-full object-cover" />
                <IconButton
                  label="Remove image"
                  onClick={() => removeImage(index)}
                  className="absolute right-2 top-2 h-8 w-8 border-white/80 bg-white/95 text-rose-600 shadow-sm hover:text-rose-700"
                >
                  <Trash2 className="h-4 w-4" aria-hidden="true" />
                </IconButton>
              </div>
              <div className="relative border-t border-slate-200">
                <LinkIcon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
                <input
                  aria-label={`Image URL ${index + 1}`}
                  value={src}
                  onChange={(event) => updateImage(index, event.target.value)}
                  className="h-10 w-full bg-white pl-9 pr-3 text-xs text-slate-700 outline-none transition focus:ring-2 focus:ring-inset focus:ring-blue-100"
                />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="rounded-md border border-dashed border-slate-200 px-3 py-6 text-center text-sm text-slate-500">
          No images
        </div>
      )}
    </section>
  );
}
