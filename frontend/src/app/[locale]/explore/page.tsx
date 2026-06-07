'use client';

import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { SearchBox } from '@/components/search-box';
import { Timeline } from '@/components/timeline';
import { api } from '@/lib/api';

export default function ExplorePage() {
  const t = useTranslations('timeline');
  const ts = useTranslations('nav');
  return (
    <AppShell>
      <PageHeader title={ts('explore')} />
      <div className="border-b border-line p-3 lg:hidden">
        <SearchBox />
      </div>
      <Timeline fetcher={(cursor) => api.globalTimeline(cursor)} emptyText={t('empty')} />
    </AppShell>
  );
}
