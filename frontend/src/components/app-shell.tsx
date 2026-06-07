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
  HomeIcon,
  MailIcon,
  LogoutIcon,
  SettingsIcon,
  UserIcon,
} from './icons';
import { YapperLogo } from './brand-logo';
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
  const { user, logout } = useSession();
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
    { href: '/settings', label: t('settings'), icon: () => <SettingsIcon className="h-7 w-7" />, authOnly: true },
  ];

  const isActive = (href: string) => (href === '/' ? pathname === '/' : pathname.startsWith(href));

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-[1280px] justify-center gap-1 px-0 sm:gap-2 sm:px-2">
      {/* Left sidebar */}
      <header className="sticky top-0 flex h-screen w-[50px] shrink-0 flex-col justify-between py-2 sm:w-[72px] sm:py-3 xl:w-[260px]">
        <div className="flex flex-col gap-0.5 sm:gap-1">
          <Link href="/" className="mb-1 flex justify-center px-1 sm:mb-2 xl:justify-start xl:px-3">
            <YapperLogo iconClassName="h-7 w-7 sm:h-9 sm:w-9" wordmarkClassName="hidden text-2xl xl:inline" />
          </Link>
          {items
            .filter((it) => !it.authOnly || user)
            .map((it) => {
              const active = isActive(it.href);
              return (
                <Link
                  key={it.href}
                  href={it.href}
                  className={`relative flex items-center justify-center gap-4 rounded-full p-2 transition hover:bg-gray-100 xl:justify-start xl:px-3 xl:py-2.5 ${
                    active ? 'font-extrabold' : 'font-normal'
                  }`}
                >
                  <span className="relative [&_svg]:h-6 [&_svg]:w-6 xl:[&_svg]:h-7 xl:[&_svg]:w-7">
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
          <div className="flex flex-col gap-2 px-1 pb-2 xl:px-2">
            <Link href={`/${user.username}`} className="flex min-w-0 items-center justify-center gap-2 rounded-full p-1 hover:bg-gray-100 xl:justify-start">
              <Avatar url={user.avatarUrl} name={user.displayName || user.username} size={32} className="xl:hidden" />
              <Avatar url={user.avatarUrl} name={user.displayName || user.username} size={40} className="hidden xl:block" />
              <span className="hidden min-w-0 xl:block">
                <span className="block truncate font-bold">{user.displayName || user.username}</span>
                <span className="block truncate text-sm text-muted">@{user.username}</span>
              </span>
            </Link>
            <button
              type="button"
              onClick={() => void logout()}
              className="rounded-full p-2 text-muted transition hover:bg-gray-100 hover:text-ink xl:px-3 xl:py-2 xl:text-left xl:text-[15px]"
              aria-label={t('logout')}
              title={t('logout')}
            >
              <span className="xl:hidden">
                <LogoutIcon className="h-6 w-6" />
              </span>
              <span className="hidden xl:inline">{t('logout')}</span>
            </button>
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
      <main className="min-h-screen min-w-0 flex-1 max-w-feed border-x border-line">{children}</main>

      {/* Right rail */}
      {rightRail && (
        <aside className="sticky top-0 hidden h-screen w-[350px] shrink-0 flex-col gap-4 overflow-y-auto py-3 pl-4 no-scrollbar lg:flex">
          <SearchBox />
          {user && <Suggestions />}
        </aside>
      )}

    </div>
  );
}
