'use client';

import { useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Link, usePathname, useRouter } from '@/i18n/navigation';
import { useDebouncedValue } from '@/hooks/use-debounced-value';
import { api } from '@/lib/api';
import type { User, Yap } from '@/lib/types';
import { Avatar } from './avatar';
import { ExploreIcon } from './icons';

const SUGGEST_LIMIT = 4;
const DEBOUNCE_MS = 300;

export function SearchBox({ initial = '' }: { initial?: string }) {
  const t = useTranslations('search');
  const router = useRouter();
  const pathname = usePathname();
  const onSearchPage = pathname === '/search';

  const [q, setQ] = useState(initial);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [yaps, setYaps] = useState<Yap[]>([]);
  const rootRef = useRef<HTMLDivElement>(null);

  const debounced = useDebouncedValue(q.trim(), DEBOUNCE_MS);

  useEffect(() => {
    setQ(initial);
  }, [initial]);

  // Sync URL while typing; wait until debounce caught up so Enter does not get overwritten.
  useEffect(() => {
    if (!onSearchPage) return;
    const typed = q.trim();
    if (typed !== debounced) return;
    const currentQ = new URL(window.location.href).searchParams.get('q') ?? '';
    if (debounced === currentQ) return;
    router.replace(debounced ? `/search?q=${encodeURIComponent(debounced)}` : '/search');
  }, [debounced, onSearchPage, q, router]);

  useEffect(() => {
    if (onSearchPage || debounced.length === 0) {
      setUsers([]);
      setYaps([]);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);

    Promise.all([
      api.searchUsers(debounced, SUGGEST_LIMIT),
      api.searchYaps(debounced, undefined, SUGGEST_LIMIT),
    ])
      .then(([u, y]) => {
        if (cancelled) return;
        setUsers(u.items ?? []);
        setYaps(y.items ?? []);
      })
      .catch(() => {
        if (cancelled) return;
        setUsers([]);
        setYaps([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [debounced, onSearchPage]);

  useEffect(() => {
    const onPointerDown = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onPointerDown);
    return () => document.removeEventListener('mousedown', onPointerDown);
  }, []);

  const goToSearch = (term: string) => {
    const trimmed = term.trim();
    if (!trimmed) return;
    setOpen(false);
    const href = `/search?q=${encodeURIComponent(trimmed)}`;
    if (onSearchPage) router.replace(href);
    else router.push(href);
  };

  const showDropdown = !onSearchPage && open && debounced.length > 0;

  return (
    <div ref={rootRef} className="sticky top-0 z-20">
      <form
        onSubmit={(e) => {
          e.preventDefault();
          goToSearch(q);
        }}
      >
        <div className="flex items-center gap-2 rounded-full bg-gray-100 px-4 py-2.5 focus-within:bg-white focus-within:ring-1 focus-within:ring-brand">
          <ExploreIcon className="h-5 w-5 shrink-0 text-muted" />
          <input
            value={q}
            onChange={(e) => {
              setQ(e.target.value);
              if (!onSearchPage) setOpen(true);
            }}
            onFocus={() => {
              if (!onSearchPage && q.trim()) setOpen(true);
            }}
            placeholder={t('placeholder')}
            className="w-full bg-transparent outline-none"
            autoComplete="off"
            spellCheck={false}
          />
        </div>
      </form>

      {showDropdown && (
        <div className="mt-2 overflow-hidden rounded-2xl border border-line bg-white shadow-lg">
          <button
            type="button"
            onClick={() => goToSearch(debounced)}
            className="flex w-full items-center gap-3 px-4 py-3 text-left hover:bg-gray-50"
          >
            <ExploreIcon className="h-5 w-5 text-muted" />
            <span>{t('searchFor', { query: debounced })}</span>
          </button>

          {users.length > 0 && (
            <div className="border-t border-line">
              <p className="px-4 py-2 text-xs font-semibold uppercase tracking-wide text-muted">{t('people')}</p>
              <ul>
                {users.map((u) => (
                  <li key={u.id}>
                    <Link
                      href={`/${u.username}`}
                      onClick={() => setOpen(false)}
                      className="flex items-center gap-3 px-4 py-2.5 hover:bg-gray-50"
                    >
                      <Avatar url={u.avatarUrl} name={u.displayName || u.username} size={40} />
                      <span className="min-w-0">
                        <span className="block truncate font-bold">{u.displayName || u.username}</span>
                        <span className="block truncate text-sm text-muted">@{u.username}</span>
                      </span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {yaps.length > 0 && (
            <div className="border-t border-line">
              <p className="px-4 py-2 text-xs font-semibold uppercase tracking-wide text-muted">{t('yaps')}</p>
              <ul>
                {yaps.map((y) => (
                  <li key={y.id}>
                    <Link
                      href={`/yap/${y.id}`}
                      onClick={() => setOpen(false)}
                      className="block px-4 py-2.5 hover:bg-gray-50"
                    >
                      <span className="block truncate text-sm font-bold">
                        {y.author.displayName || y.author.username}
                      </span>
                      <span className="block truncate text-sm text-muted">{y.content}</span>
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {loading && users.length === 0 && yaps.length === 0 ? (
            <p className="border-t border-line px-4 py-3 text-sm text-muted">{t('searching')}</p>
          ) : !loading && users.length === 0 && yaps.length === 0 ? (
            <p className="border-t border-line px-4 py-3 text-sm text-muted">{t('noResults')}</p>
          ) : null}
        </div>
      )}
    </div>
  );
}
