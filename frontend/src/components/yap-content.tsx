'use client';

import { Link } from '@/i18n/navigation';
import { tokenize } from '@/lib/format';

export function YapContent({ content }: { content: string }) {
  const segments = tokenize(content);
  return (
    <span className="whitespace-pre-wrap break-words">
      {segments.map((seg, i) => {
        switch (seg.type) {
          case 'hashtag':
            return (
              <Link key={i} href={`/hashtag/${encodeURIComponent(seg.value.slice(1))}`} className="text-brand hover:underline" onClick={(e) => e.stopPropagation()}>
                {seg.value}
              </Link>
            );
          case 'mention':
            return (
              <Link key={i} href={`/${seg.value.slice(1)}`} className="text-brand hover:underline" onClick={(e) => e.stopPropagation()}>
                {seg.value}
              </Link>
            );
          case 'url':
            return (
              <a key={i} href={seg.value} target="_blank" rel="noreferrer" className="text-brand hover:underline" onClick={(e) => e.stopPropagation()}>
                {seg.value}
              </a>
            );
          default:
            return <span key={i}>{seg.value}</span>;
        }
      })}
    </span>
  );
}
