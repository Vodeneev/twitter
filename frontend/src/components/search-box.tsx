'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/navigation';
import { ExploreIcon } from './icons';

export function SearchBox({ initial = '' }: { initial?: string }) {
  const t = useTranslations('search');
  const router = useRouter();
  const [q, setQ] = useState(initial);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        const term = q.trim();
        if (term) router.push(`/search?q=${encodeURIComponent(term)}`);
      }}
      className="sticky top-0"
    >
      <div className="flex items-center gap-2 rounded-full bg-gray-100 px-4 py-2.5 focus-within:bg-white focus-within:ring-1 focus-within:ring-brand">
        <ExploreIcon className="h-5 w-5 text-muted" />
        <input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder={t('placeholder')}
          className="w-full bg-transparent outline-none"
        />
      </div>
    </form>
  );
}
