'use client';

import { useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { Timeline } from '@/components/timeline';
import { RequireAuth } from '@/components/require-auth';
import { api } from '@/lib/api';

export default function BookmarksPage() {
  const t = useTranslations('bookmarks');
  return (
    <AppShell>
      <PageHeader title={t('title')} />
      <RequireAuth>
        <Timeline fetcher={(cursor) => api.bookmarks(cursor)} emptyText={t('empty')} />
      </RequireAuth>
    </AppShell>
  );
}
