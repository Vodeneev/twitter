'use client';

import { Suspense, useEffect, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AuthCard } from '@/components/auth-card';
import { Link } from '@/i18n/navigation';
import { api } from '@/lib/api';

export const dynamic = 'force-dynamic';

function VerifyEmailInner() {
  const t = useTranslations('auth');
  const params = useSearchParams();
  const token = params.get('token') ?? '';
  const [status, setStatus] = useState<'pending' | 'ok' | 'fail'>('pending');

  useEffect(() => {
    if (!token) {
      setStatus('fail');
      return;
    }
    api
      .verifyEmail(token)
      .then(() => setStatus('ok'))
      .catch(() => setStatus('fail'));
  }, [token]);

  return (
    <AuthCard title={t('loginTitle')}>
      {status === 'pending' && <p className="text-center text-muted">{t('verifying')}</p>}
      {status === 'ok' && <p className="alert-success">{t('verified')}</p>}
      {status === 'fail' && <p className="alert-error text-center">{t('verifyFailed')}</p>}
      <div className="mt-4 text-center text-sm">
        <Link href="/login" className="text-brand hover:underline">
          {t('loginButton')}
        </Link>
      </div>
    </AuthCard>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={null}>
      <VerifyEmailInner />
    </Suspense>
  );
}
