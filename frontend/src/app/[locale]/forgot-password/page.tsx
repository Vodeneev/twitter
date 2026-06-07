'use client';

import { useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { AuthCard } from '@/components/auth-card';
import { Link } from '@/i18n/navigation';
import { api } from '@/lib/api';

export default function ForgotPasswordPage() {
  const t = useTranslations('auth');
  const locale = useLocale();
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    try {
      await api.forgotPassword(email, locale);
      setSent(true);
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthCard title={t('forgotTitle')}>
      {sent ? (
        <p className="alert-info">{t('resetSent')}</p>
      ) : (
        <form onSubmit={submit} className="flex flex-col gap-4">
          <p className="text-sm text-muted">{t('forgotHint')}</p>
          <input className="input" type="email" placeholder={t('email')} value={email} onChange={(e) => setEmail(e.target.value)} required />
          <button type="submit" disabled={busy} className="btn-primary">
            {t('sendReset')}
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
