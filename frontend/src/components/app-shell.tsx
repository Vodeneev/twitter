'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
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
  LoginIcon,
  LogoutIcon,
  MailIcon,
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
  /** Hide in the icon rail on wide desktop — shown elsewhere (e.g. footer buttons). */
  hideOnXl?: boolean;
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

  const items: NavItem[] = useMemo(() => {
    const base: NavItem[] = [
      { href: '/', label: t('home'), icon: (a) => <HomeIcon className="h-7 w-7" filled={a} /> },
      { href: '/search', label: t('search'), icon: () => <ExploreIcon className="h-7 w-7" /> },
    ];
    if (!user) {
      return [
        ...base,
        { href: '/login', label: t('login'), icon: () => <LoginIcon className="h-7 w-7" />, hideOnXl: true },
        { href: '/register', label: t('register'), icon: () => <UserIcon className="h-7 w-7" />, hideOnXl: true },
      ];
    }
    return [
      ...base,
      { href: '/notifications', label: t('notifications'), icon: (a) => <BellIcon className="h-7 w-7" filled={a} />, badge: unreadNotif },
      { href: '/messages', label: t('messages'), icon: (a) => <MailIcon className="h-7 w-7" filled={a} />, badge: unreadMsg },
      { href: '/bookmarks', label: t('bookmarks'), icon: (a) => <BookmarkIcon className="h-7 w-7" filled={a} /> },
      { href: `/${user.username}`, label: t('profile'), icon: (a) => <UserIcon className="h-7 w-7" filled={a} /> },
      { href: '/settings', label: t('settings'), icon: () => <SettingsIcon className="h-7 w-7" /> },
    ];
  }, [t, user, unreadNotif, unreadMsg]);

  const isActive = (href: string) => (href === '/' ? pathname === '/' : pathname.startsWith(href));

  const footerPad = 'pb-[max(0.75rem,env(safe-area-inset-bottom))]';

  return (
    <div className="mx-auto flex min-h-screen w-full max-w-[1280px] justify-center gap-1 px-0 sm:gap-2 sm:px-2">
      {/* Left sidebar */}
      <header className="sticky top-0 flex h-dvh max-h-dvh w-[50px] shrink-0 flex-col py-2 sm:w-[72px] sm:py-3 xl:w-[260px]">
        <div className="flex min-h-0 flex-1 flex-col gap-0.5 overflow-y-auto sm:gap-1">
          <Link href="/" className="mb-1 flex shrink-0 justify-center px-1 sm:mb-2 xl:justify-start xl:px-3">
            <YapperLogo iconClassName="h-7 w-7 sm:h-9 sm:w-9" wordmarkClassName="hidden text-2xl xl:inline" />
          </Link>
          {items.map((it) => {
            const active = isActive(it.href);
            return (
              <Link
                key={it.href}
                href={it.href}
                className={`relative flex shrink-0 items-center justify-center gap-4 rounded-full p-2 transition hover:bg-gray-100 xl:justify-start xl:px-3 xl:py-2.5 ${
                  active ? 'font-extrabold' : 'font-normal'
                } ${it.hideOnXl ? 'xl:hidden' : ''}`}
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
            <Link href="/" className="btn-primary mt-3 hidden h-12 shrink-0 text-lg xl:flex">
              {tb('name')}
            </Link>
          )}
        </div>

        {user ? (
          <div className={`mt-2 flex shrink-0 flex-col gap-2 border-t border-line px-1 pt-2 xl:px-2 ${footerPad}`}>
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
              className="flex items-center justify-center rounded-full p-2 text-muted transition hover:bg-gray-100 hover:text-ink xl:justify-start xl:px-3 xl:py-2"
              aria-label={t('logout')}
              title={t('logout')}
            >
              <LogoutIcon className="h-6 w-6 xl:mr-0" />
              <span className="hidden xl:ml-3 xl:inline xl:text-[15px]">{t('logout')}</span>
            </button>
          </div>
        ) : (
          <div className={`mt-2 hidden shrink-0 flex-col gap-2 border-t border-line px-2 pt-2 xl:flex ${footerPad}`}>
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
