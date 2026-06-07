'use client';

import { useTranslations } from 'next-intl';
import { Link } from '@/i18n/navigation';
import { FeatherIcon } from './icons';

export function AuthCard({ title, children }: { title: string; children: React.ReactNode }) {
  const tb = useTranslations('brand');
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md rounded-2xl border border-line bg-white p-8 shadow-sm">
        <Link href="/" className="mb-4 flex items-center justify-center gap-2 text-brand">
          <FeatherIcon className="h-10 w-10" />
          <span className="text-2xl font-extrabold text-ink">{tb('name')}</span>
        </Link>
        <h1 className="mb-6 text-center text-2xl font-extrabold">{title}</h1>
        {children}
      </div>
    </div>
  );
}
