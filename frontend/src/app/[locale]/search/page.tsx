'use client';

import { Suspense, useCallback, useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader, Tabs } from '@/components/page-header';
import { SearchBox } from '@/components/search-box';
import { Timeline } from '@/components/timeline';
import { UserList } from '@/components/user-list';
import { api } from '@/lib/api';
import type { User } from '@/lib/types';

export const dynamic = 'force-dynamic';

type Tab = 'yaps' | 'people';

function SearchInner() {
  const t = useTranslations('search');
  const params = useSearchParams();
  const q = params.get('q') ?? '';
  const [tab, setTab] = useState<Tab>('yaps');
  const [people, setPeople] = useState<User[]>([]);

  useEffect(() => {
    if (!q) return;
    api.searchUsers(q).then((r) => setPeople(r.items ?? [])).catch(() => setPeople([]));
  }, [q]);

  const fetcher = useCallback((cursor?: string) => api.searchYaps(q, cursor), [q]);

  return (
    <AppShell>
      <PageHeader title={t('title')} subtitle={q ? `“${q}”` : undefined} back />
      <div className="border-b border-line p-3">
        <SearchBox initial={q} />
      </div>
      {!q ? (
        <div className="px-6 py-12 text-center text-muted">{t('typeToSearch')}</div>
      ) : (
        <>
          <Tabs<Tab>
            tabs={[
              { id: 'yaps', label: t('yaps') },
              { id: 'people', label: t('people') },
            ]}
            active={tab}
            onChange={setTab}
          />
          {tab === 'people' ? (
            <UserList users={people} empty={t('noResults')} />
          ) : (
            <Timeline key={q} fetcher={fetcher} emptyText={t('noResults')} />
          )}
        </>
      )}
    </AppShell>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={null}>
      <SearchInner />
    </Suspense>
  );
}
