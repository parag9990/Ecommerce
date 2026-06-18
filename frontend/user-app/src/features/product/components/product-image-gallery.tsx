import { useState } from 'react';

type ProductImageGalleryProps = {
  images?: string[] | undefined;
  title: string;
};

const emptyImages: string[] = [];

export function ProductImageGallery({
  images,
  title,
}: ProductImageGalleryProps) {
  const galleryImages = images ?? emptyImages;
  const firstImage = galleryImages[0];
  const [selection, setSelection] = useState<{
    image?: string | undefined;
    sourceImage?: string | undefined;
  }>({
    image: firstImage,
    sourceImage: firstImage,
  });
  const selectedImage =
    selection.sourceImage === firstImage &&
    selection.image &&
    galleryImages.includes(selection.image)
      ? selection.image
      : firstImage;

  return (
    <section aria-label={`${title} images`} className="space-y-3">
      <div className="aspect-square overflow-hidden rounded-md border border-slate-200 bg-slate-100">
        {selectedImage ? (
          <img
            alt={title}
            className="h-full w-full object-cover"
            src={selectedImage}
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center px-6 text-center text-sm text-slate-500">
            Product image coming soon
          </div>
        )}
      </div>

      {galleryImages.length > 1 ? (
        <div className="grid grid-cols-4 gap-2">
          {galleryImages.slice(0, 4).map((image) => (
            <button
              aria-label={`Show ${title} image`}
              className={[
                'aspect-square overflow-hidden rounded-md border bg-white',
                selectedImage === image
                  ? 'border-slate-950'
                  : 'border-slate-200 hover:border-slate-400',
              ].join(' ')}
              key={image}
              onClick={() => {
                setSelection({
                  image,
                  sourceImage: firstImage,
                });
              }}
              type="button"
            >
              <img
                alt=""
                className="h-full w-full object-cover"
                loading="lazy"
                src={image}
              />
            </button>
          ))}
        </div>
      ) : null}
    </section>
  );
}
