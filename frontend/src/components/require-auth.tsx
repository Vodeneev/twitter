'use client';

import { useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useSession } from './session-provider';
import { useRouter } from '@/i18n/navigation';

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useSession();
  const router = useRouter();
  const t = useTranslations('common');

  useEffect(() => {
    if (!loading && !user) router.replace('/login');
  }, [loading, user, router]);

  if (loading) return <div className="py-10 text-center text-muted">{t('loading')}</div>;
  if (!user) return null;
  return <>{children}</>;
}
