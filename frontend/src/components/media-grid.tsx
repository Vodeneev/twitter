'use client';

import type { Media } from '@/lib/types';

export function MediaGrid({ media }: { media: Media[] }) {
  if (!media || media.length === 0) return null;
  const count = media.length;
  const gridCls = count === 1 ? 'grid-cols-1' : 'grid-cols-2';
  return (
    <div className={`mt-3 grid ${gridCls} gap-0.5 overflow-hidden rounded-2xl border border-line`}>
      {media.map((m) => (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          key={m.id}
          src={m.url}
          alt=""
          className={`w-full object-cover ${count === 1 ? 'max-h-[510px]' : 'h-44 sm:h-60'}`}
          onClick={(e) => e.stopPropagation()}
        />
      ))}
    </div>
  );
}
