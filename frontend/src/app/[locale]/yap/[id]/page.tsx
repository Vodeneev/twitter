'use client';

import { useCallback, useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { Timeline } from '@/components/timeline';
import { YapCard } from '@/components/yap-card';
import { YapComposer } from '@/components/yap-composer';
import { useSession } from '@/components/session-provider';
import { api } from '@/lib/api';
import type { Yap } from '@/lib/types';

export default function YapDetailPage() {
  const t = useTranslations('yap');
  const params = useParams();
  const id = String(params.id);
  const { user } = useSession();
  const [yap, setYap] = useState<Yap | null>(null);
  const [ancestors, setAncestors] = useState<Yap[]>([]);
  const [head, setHead] = useState<Yap[]>([]);

  useEffect(() => {
    setHead([]);
    api
      .thread(id)
      .then((r) => {
        setYap(r.yap);
        setAncestors(r.ancestors ?? []);
      })
      .catch(() => setYap(null));
  }, [id]);

  const fetcher = useCallback((cursor?: string) => api.replies(id, cursor), [id]);

  return (
    <AppShell>
      <PageHeader title={t('replies')} back />
      {ancestors.map((a) => (
        <YapCard key={a.id} yap={a} showThreadLine />
      ))}
      {yap && <YapCard yap={yap} />}
      {user && (
        <div className="border-b border-line">
          <YapComposer replyToId={id} placeholder={t('reply')} onCreated={(y) => setHead((h) => [y, ...h])} />
        </div>
      )}
      <Timeline fetcher={fetcher} emptyText="—" headItems={head} />
    </AppShell>
  );
}
