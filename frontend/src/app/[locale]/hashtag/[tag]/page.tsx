'use client';

import { useCallback } from 'react';
import { useParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { Timeline } from '@/components/timeline';
import { api } from '@/lib/api';

export default function HashtagPage() {
  const t = useTranslations('timeline');
  const params = useParams();
  const tag = decodeURIComponent(String(params.tag));
  const fetcher = useCallback((cursor?: string) => api.hashtag(tag, cursor), [tag]);

  return (
    <AppShell>
      <PageHeader title={`#${tag}`} back />
      <Timeline fetcher={fetcher} emptyText={t('empty')} />
    </AppShell>
  );
}
