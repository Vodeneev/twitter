'use client';

import { useCallback, useState } from 'react';
import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { Tabs } from '@/components/page-header';
import { Timeline } from '@/components/timeline';
import { YapComposer } from '@/components/yap-composer';
import { useSession } from '@/components/session-provider';
import { api } from '@/lib/api';
import type { Yap } from '@/lib/types';

type Tab = 'forYou' | 'following';

export default function HomePage() {
  const t = useTranslations('timeline');
  const { user } = useSession();
  const [tab, setTab] = useState<Tab>('forYou');
  const [head, setHead] = useState<Yap[]>([]);

  const fetcher = useCallback(
    (cursor?: string) => (tab === 'following' && user ? api.homeTimeline(cursor) : api.globalTimeline(cursor)),
    [tab, user],
  );

  return (
    <AppShell>
      {user ? (
        <Tabs<Tab>
          tabs={[
            { id: 'forYou', label: t('forYou') },
            { id: 'following', label: t('following') },
          ]}
          active={tab}
          onChange={(id) => {
            setTab(id);
            setHead([]);
          }}
        />
      ) : (
        <div className="sticky top-0 z-10 border-b border-line bg-white/80 px-4 py-4 backdrop-blur">
          <h1 className="text-xl font-extrabold">{t('forYou')}</h1>
        </div>
      )}

      {user && (
        <div className="border-b border-line">
          <YapComposer onCreated={(yap) => setHead((h) => [yap, ...h])} />
        </div>
      )}

      <Timeline fetcher={fetcher} emptyText={t('empty')} reloadToken={tab === 'following' ? 1 : 0} headItems={tab === 'following' ? head : head} />
    </AppShell>
  );
}
