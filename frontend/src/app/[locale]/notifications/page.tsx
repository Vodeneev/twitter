'use client';

import { useCallback, useEffect, useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { RequireAuth } from '@/components/require-auth';
import { Avatar } from '@/components/avatar';
import { Link } from '@/i18n/navigation';
import { useRealtime, type RealtimeEvent } from '@/hooks/use-realtime';
import { useSession } from '@/components/session-provider';
import { api } from '@/lib/api';
import { timeAgo } from '@/lib/format';
import { BellIcon, HeartIcon, RepostIcon, ReplyIcon, UserIcon, QuoteIcon } from '@/components/icons';
import type { AppNotification, NotificationType } from '@/lib/types';

function icon(type: NotificationType) {
  switch (type) {
    case 'like':
      return <HeartIcon className="h-6 w-6 text-pink-600" filled />;
    case 'follow':
      return <UserIcon className="h-6 w-6 text-brand" filled />;
    case 'repost':
      return <RepostIcon className="h-6 w-6 text-green-600" />;
    case 'reply':
      return <ReplyIcon className="h-6 w-6 text-brand" />;
    case 'quote':
      return <QuoteIcon className="h-6 w-6 text-brand" />;
    default:
      return <BellIcon className="h-6 w-6 text-brand" />;
  }
}

function NotificationsInner() {
  const t = useTranslations('notifications');
  const locale = useLocale();
  const { user } = useSession();
  const [items, setItems] = useState<AppNotification[]>([]);
  const [cursor, setCursor] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    const page = await api.notifications();
    setItems(page.items ?? []);
    setCursor(page.nextCursor ?? undefined);
    setLoading(false);
    await api.markNotificationsRead().catch(() => {});
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const onEvent = useCallback((ev: RealtimeEvent) => {
    if (ev.type === 'notification') {
      setItems((prev) => [ev.data as AppNotification, ...prev]);
    }
  }, []);
  useRealtime(Boolean(user), onEvent);

  const loadMore = async () => {
    if (!cursor) return;
    const page = await api.notifications(cursor);
    setItems((prev) => [...prev, ...(page.items ?? [])]);
    setCursor(page.nextCursor ?? undefined);
  };

  if (loading) return <div className="py-10 text-center text-muted">…</div>;
  if (items.length === 0) return <div className="px-6 py-12 text-center text-muted">{t('empty')}</div>;

  return (
    <div>
      {items.map((n) => {
        const name = n.actor.displayName || n.actor.username;
        const href = n.yapId ? `/yap/${n.yapId}` : `/${n.actor.username}`;
        return (
          <Link key={n.id} href={href} className={`flex gap-3 border-b border-line px-4 py-3 hover:bg-gray-50 ${n.read ? '' : 'bg-ya-yellow/10'}`}>
            <div className="pt-1">{icon(n.type)}</div>
            <div className="min-w-0">
              <Avatar url={n.actor.avatarUrl} name={name} size={32} />
              <p className="mt-1">
                <span className="font-bold">{name}</span> <span className="text-muted">{t(n.type)}</span>{' '}
                <span className="text-muted">· {timeAgo(n.createdAt, locale)}</span>
              </p>
              {n.yapPreview && <p className="mt-0.5 text-muted">{n.yapPreview}</p>}
            </div>
          </Link>
        );
      })}
      {cursor && (
        <button onClick={loadMore} className="w-full py-4 text-center font-semibold text-brand hover:bg-gray-50">
          ↓
        </button>
      )}
    </div>
  );
}

export default function NotificationsPage() {
  const t = useTranslations('notifications');
  return (
    <AppShell>
      <PageHeader title={t('title')} />
      <RequireAuth>
        <NotificationsInner />
      </RequireAuth>
    </AppShell>
  );
}
