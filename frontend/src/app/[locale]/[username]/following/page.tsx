'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { UserList } from '@/components/user-list';
import { api } from '@/lib/api';
import type { User } from '@/lib/types';

export default function FollowingPage() {
  const t = useTranslations('profile');
  const params = useParams();
  const username = decodeURIComponent(String(params.username));
  const [users, setUsers] = useState<User[]>([]);

  useEffect(() => {
    api.followingList(username).then((r) => setUsers(r.items ?? [])).catch(() => setUsers([]));
  }, [username]);

  return (
    <AppShell>
      <PageHeader title={t('followingCount')} subtitle={`@${username}`} back />
      <UserList users={users} empty="—" />
    </AppShell>
  );
}
