'use client';

import type { User } from '@/lib/types';
import { Link } from '@/i18n/navigation';
import { Avatar } from './avatar';
import { FollowButton } from './follow-button';

export function UserList({ users, empty }: { users: User[]; empty: string }) {
  if (users.length === 0) {
    return <div className="px-6 py-12 text-center text-muted">{empty}</div>;
  }
  return (
    <ul>
      {users.map((u) => (
        <li key={u.id} className="flex items-start justify-between gap-3 border-b border-line px-4 py-3 hover:bg-gray-50">
          <Link href={`/${u.username}`} className="flex min-w-0 gap-3">
            <Avatar url={u.avatarUrl} name={u.displayName || u.username} size={48} />
            <span className="min-w-0">
              <span className="block truncate font-bold">{u.displayName || u.username}</span>
              <span className="block truncate text-muted">@{u.username}</span>
              {u.bio && <span className="mt-0.5 block line-clamp-2 text-sm">{u.bio}</span>}
            </span>
          </Link>
          <div className="shrink-0">
            <FollowButton username={u.username} initialFollowing={u.following} small />
          </div>
        </li>
      ))}
    </ul>
  );
}
