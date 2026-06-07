'use client';

import { useEffect, useState } from 'react';
import { useTranslations } from 'next-intl';
import { api } from '@/lib/api';
import type { User } from '@/lib/types';
import { Link } from '@/i18n/navigation';
import { Avatar } from './avatar';
import { FollowButton } from './follow-button';

export function Suggestions() {
  const t = useTranslations('suggestions');
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    api
      .suggestions(5)
      .then((r) => setUsers(r.items ?? []))
      .catch(() => setUsers([]));
  }, []);

  if (users.length === 0) return null;

  return (
    <div className="rounded-2xl bg-gray-50">
      <h2 className="px-4 pt-3 text-xl font-extrabold">{t('title')}</h2>
      <ul>
        {users.map((u) => (
          <li key={u.id} className="flex items-center justify-between gap-2 px-4 py-3 hover:bg-gray-100">
            <Link href={`/${u.username}`} className="flex min-w-0 items-center gap-2">
              <Avatar url={u.avatarUrl} name={u.displayName || u.username} size={40} />
              <span className="min-w-0">
                <span className="block truncate font-bold">{u.displayName || u.username}</span>
                <span className="block truncate text-sm text-muted">@{u.username}</span>
              </span>
            </Link>
            <FollowButton username={u.username} initialFollowing={u.following} small />
          </li>
        ))}
      </ul>
    </div>
  );
}
