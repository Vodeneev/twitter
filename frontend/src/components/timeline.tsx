'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import type { Page, Yap } from '@/lib/types';
import { YapCard } from './yap-card';

interface Props {
  fetcher: (cursor?: string) => Promise<Page<Yap>>;
  emptyText: string;
  reloadToken?: number;
  headItems?: Yap[];
}

export function Timeline({ fetcher, emptyText, reloadToken = 0, headItems = [] }: Props) {
  const t = useTranslations('timeline');
  const [items, setItems] = useState<Yap[]>([]);
  const [cursor, setCursor] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;

  const loadFirst = useCallback(async () => {
    setLoading(true);
    try {
      const page = await fetcherRef.current();
      setItems(page.items ?? []);
      setCursor(page.nextCursor ?? undefined);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadFirst();
  }, [loadFirst, reloadToken]);

  const loadMore = async () => {
    if (!cursor || loadingMore) return;
    setLoadingMore(true);
    try {
      const page = await fetcherRef.current(cursor);
      setItems((prev) => [...prev, ...(page.items ?? [])]);
      setCursor(page.nextCursor ?? undefined);
    } finally {
      setLoadingMore(false);
    }
  };

  const onChange = (next: Yap) => setItems((prev) => prev.map((y) => (y.id === next.id ? next : y)));
  const onDelete = (id: string) => setItems((prev) => prev.filter((y) => y.id !== id));

  const seen = new Set(items.map((y) => y.id));
  const head = headItems.filter((y) => !seen.has(y.id));
  const all = [...head, ...items];

  if (loading) {
    return <div className="py-10 text-center text-muted">{t('loading')}</div>;
  }
  if (all.length === 0) {
    return <div className="px-6 py-12 text-center text-muted">{emptyText}</div>;
  }

  return (
    <div>
      {all.map((y) => (
        <YapCard key={`${y.id}-${y.repostedAt ?? ''}`} yap={y} onChange={onChange} onDelete={onDelete} />
      ))}
      {cursor && (
        <button onClick={loadMore} disabled={loadingMore} className="w-full py-4 text-center font-semibold text-brand hover:bg-gray-50">
          {loadingMore ? t('loading') : t('loadMore')}
        </button>
      )}
    </div>
  );
}
