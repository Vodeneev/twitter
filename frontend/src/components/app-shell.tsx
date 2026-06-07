'use client';

import { useCallback, useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { Link, usePathname } from '@/i18n/navigation';
import { useSession } from './session-provider';
import { useRealtime, type RealtimeEvent } from '@/hooks/use-realtime';
import { api } from '@/lib/api';
import { Avatar } from './avatar';
import {
  BellIcon,
  BookmarkIcon,
  ExploreIcon,
  FeatherIcon,
  HomeIcon,
  MailIcon,
  UserIcon,
} from './icons';
import { Suggestions } from './suggestions';
import { SearchBox } from './search-box';

interface NavItem {
  href: string;
  label: string;
  icon: (active: boolean) => React.ReactNode;
  badge?: number;
  authOnly?: boolean;
}

export function AppShell({ children, rightRail = true }: { children: React.ReactNode; rightRail?: boolean }) {
  const t = useTranslations('nav');
  const tb = useTranslations('brand');
  const pathname = usePathname();
  const { user } = useSession();
  const [unreadNotif, setUnreadNotif] = useState(0);
  const [unreadMsg, setUnreadMsg] = useState(0);

  const loadCounts = useCallback(async () => {
    if (!user) return;
    try {
      const [{ count }, { items }] = await Promise.all([api.unreadCount(), api.conversations()]);
      setUnreadNotif(count);
      setUnreadMsg(items.reduce((acc, c) => acc + (c.unread > 0 ? 1 : 0), 0));
    } catch {
      /* ignore */
    }
  }, [user]);

  useEffect(() => {
    void loadCounts();
  }, [loadCounts]);

  const onEvent = useCallback(
    (ev: RealtimeEvent) => {
      if (ev.type === 'notification' && !pathname.startsWith('/notifications')) {
        setUnreadNotif((n) => n + 1);
      }
      if (ev.type === 'message' && !pathname.startsWith('/messages')) {
        setUnreadMsg((n) => n + 1);
      }
    },
    [pathname],
  );
  useRealtime(Boolean(user), onEvent);

  useEffect(() => {
    if (pathname.startsWith('/notifications')) setUnreadNotif(0);
    if (pathname.startsWith('/messages')) void loadCounts();
  }, [pathname, loadCounts]);

  const items: NavItem[] = [
    { href: '/', label: t('home'), icon: (a) => <HomeIcon className="h-7 w-7" filled={a} /> },
    { href: '/search', label: t('search'), icon: () => <ExploreIcon className="h-7 w-7" /> },
    { href: '/notifications', label: t('notifications'), icon: (a) => <BellIcon className="h-7 w-7" filled={a} />, badge: unreadNotif, authOnly: true },
    { href: '/messages', label: t('messages'), icon: (a) => <MailIcon className="h-7 w-7" filled={a} />, badge: unreadMsg, authOnly: true },
    { href: '/bookmarks', label: t('bookmarks'), icon: (a) => <BookmarkIcon className="h-7 w-7" filled={a} />, authOnly: true },
    { href: user ? `/${user.username}` : '/login', label: t('profile'), icon: (a) => <UserIcon className="h-7 w-7" filled={a} />, authOnly: true },
  ];

  const isActive = (href: string) => (href === '/' ? pathname === '/' : pathname.startsWith(href));

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-[1280px] justify-center gap-2 px-2">
      {/* Left sidebar */}
      <header className="sticky top-0 hidden h-screen shrink-0 flex-col justify-between py-3 sm:flex sm:w-[88px] xl:w-[260px]">
        <div className="flex flex-col gap-1">
          <Link href="/" className="mb-2 flex items-center gap-2 px-3 text-brand">
            <FeatherIcon className="h-9 w-9" />
            <span className="hidden text-2xl font-extrabold text-ink xl:inline">{tb('name')}</span>
          </Link>
          {items
            .filter((it) => !it.authOnly || user)
            .map((it) => {
              const active = isActive(it.href);
              return (
                <Link
                  key={it.href}
                  href={it.href}
                  className={`relative flex items-center gap-4 rounded-full px-3 py-2.5 text-xl transition hover:bg-gray-100 ${
                    active ? 'font-extrabold' : 'font-normal'
                  }`}
                >
                  <span className="relative">
                    {it.icon(active)}
                    {it.badge ? (
                      <span className="absolute -right-1.5 -top-1.5 min-w-[18px] rounded-full bg-brand px-1 text-center text-[11px] font-bold leading-[18px] text-white">
                        {it.badge > 99 ? '99+' : it.badge}
                      </span>
                    ) : null}
                  </span>
                  <span className="hidden xl:inline">{it.label}</span>
                </Link>
              );
            })}
          {user && (
            <Link href="/" className="btn-primary mt-3 hidden h-12 text-lg xl:flex">
              {tb('name')}
            </Link>
          )}
        </div>

        {user ? (
          <div className="flex items-center justify-between gap-2 px-2 pb-2">
            <Link href={`/${user.username}`} className="flex min-w-0 items-center gap-2">
              <Avatar url={user.avatarUrl} name={user.displayName || user.username} size={40} />
              <span className="hidden min-w-0 xl:block">
                <span className="block truncate font-bold">{user.displayName || user.username}</span>
                <span className="block truncate text-sm text-muted">@{user.username}</span>
              </span>
            </Link>
          </div>
        ) : (
          <div className="hidden flex-col gap-2 px-2 pb-2 xl:flex">
            <Link href="/login" className="btn-outline">
              {t('login')}
            </Link>
            <Link href="/register" className="btn-primary">
              {t('register')}
            </Link>
          </div>
        )}
      </header>

      {/* Main column */}
      <main className="min-h-screen w-full max-w-feed border-x border-line">{children}</main>

      {/* Right rail */}
      {rightRail && (
        <aside className="sticky top-0 hidden h-screen w-[350px] shrink-0 flex-col gap-4 overflow-y-auto py-3 pl-4 no-scrollbar lg:flex">
          <SearchBox />
          {user && <Suggestions />}
        </aside>
      )}

      {/* Mobile bottom nav */}
      <nav className="fixed bottom-0 left-0 right-0 z-20 flex items-center justify-around border-t border-line bg-white py-2 sm:hidden">
        {items
          .filter((it) => !it.authOnly || user)
          .map((it) => (
            <Link key={it.href} href={it.href} className="relative p-2">
              {it.icon(isActive(it.href))}
              {it.badge ? <span className="absolute right-0 top-0 h-2 w-2 rounded-full bg-brand" /> : null}
            </Link>
          ))}
      </nav>
    </div>
  );
}
