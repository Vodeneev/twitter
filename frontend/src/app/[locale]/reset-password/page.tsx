'use client';

import { Suspense, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { AuthCard } from '@/components/auth-card';
import { Link } from '@/i18n/navigation';
import { api, ApiError } from '@/lib/api';

export const dynamic = 'force-dynamic';

function ResetPasswordInner() {
  const t = useTranslations('auth');
  const te = useTranslations('auth.errors');
  const params = useSearchParams();
  const token = params.get('token') ?? '';
  const [password, setPassword] = useState('');
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      await api.resetPassword(token, password);
      setDone(true);
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation') setError(err.message);
      else if (err instanceof ApiError && err.code === 'invalid_token') setError(t('verifyFailed'));
      else setError(te('internal_error'));
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthCard title={t('resetTitle')}>
      {done ? (
        <p className="rounded-md bg-green-50 px-3 py-3 text-center text-green-700">{t('resetDone')}</p>
      ) : (
        <form onSubmit={submit} className="flex flex-col gap-4">
          {error && <p className="alert-error">{error}</p>}
          <input className="input" type="password" placeholder={t('newPassword')} value={password} onChange={(e) => setPassword(e.target.value)} minLength={8} required />
          <button type="submit" disabled={busy || !token} className="btn-primary">
            {t('resetButton')}
          </button>
        </form>
      )}
      <div className="mt-4 text-center text-sm">
        <Link href="/login" className="text-brand hover:underline">
          {t('loginButton')}
        </Link>
      </div>
    </AuthCard>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={null}>
      <ResetPasswordInner />
    </Suspense>
  );
}
