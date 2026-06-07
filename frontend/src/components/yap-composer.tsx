'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { api, uploadToStorage } from '@/lib/api';
import type { Yap } from '@/lib/types';
import { useSession } from './session-provider';
import { Avatar } from './avatar';
import { ImageIcon } from './icons';

const MAX = 280;

interface Props {
  replyToId?: string;
  quoteOfId?: string;
  placeholder?: string;
  autoFocus?: boolean;
  onCreated?: (yap: Yap) => void;
}

interface Draft {
  file: File;
  previewUrl: string;
}

export function YapComposer({ replyToId, quoteOfId, placeholder, autoFocus, onCreated }: Props) {
  const t = useTranslations('composer');
  const { user } = useSession();
  const [content, setContent] = useState('');
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [busy, setBusy] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const resizeTextarea = useCallback(() => {
    const el = textareaRef.current;
    if (!el) return;
    el.style.height = 'auto';
    el.style.height = `${el.scrollHeight}px`;
  }, []);

  useEffect(() => {
    resizeTextarea();
  }, [content, resizeTextarea]);

  if (!user) return null;

  const remaining = MAX - content.length;
  const tooLong = remaining < 0;
  const canSubmit = (content.trim().length > 0 || drafts.length > 0) && !tooLong && !busy;

  const addFiles = (files: FileList | null) => {
    if (!files) return;
    const next = Array.from(files)
      .slice(0, 4 - drafts.length)
      .map((file) => ({ file, previewUrl: URL.createObjectURL(file) }));
    setDrafts((d) => [...d, ...next].slice(0, 4));
  };

  const submit = async () => {
    if (!canSubmit) return;
    setBusy(true);
    try {
      const mediaKeys: string[] = [];
      for (const d of drafts) {
        const { uploadUrl, key } = await api.presign(d.file.type, 'yap');
        await uploadToStorage(uploadUrl, d.file);
        mediaKeys.push(key);
      }
      const { yap } = await api.createYap({
        content: content.trim(),
        replyToId,
        quoteOfId,
        mediaKeys,
      });
      setContent('');
      setDrafts([]);
      requestAnimationFrame(resizeTextarea);
      onCreated?.(yap);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex gap-3 px-4 py-3">
      <Avatar url={user.avatarUrl} name={user.displayName || user.username} size={44} />
      <div className="flex-1">
        <textarea
          ref={textareaRef}
          autoFocus={autoFocus}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={placeholder ?? t('placeholder')}
          rows={replyToId ? 3 : 4}
          className="max-h-72 min-h-[6.5rem] w-full resize-none overflow-y-auto border-0 bg-transparent text-xl leading-relaxed outline-none placeholder:text-muted"
        />
        {drafts.length > 0 && (
          <div className="mb-2 grid grid-cols-2 gap-1">
            {drafts.map((d, i) => (
              <div key={i} className="relative">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img src={d.previewUrl} alt="" className="h-40 w-full rounded-xl object-cover" />
                <button
                  type="button"
                  onClick={() => setDrafts((ds) => ds.filter((_, j) => j !== i))}
                  className="absolute right-1 top-1 rounded-full bg-black/60 px-2 text-white"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="flex items-center justify-between border-t border-line pt-2">
          <button type="button" onClick={() => fileRef.current?.click()} className="rounded-full p-2 text-brand hover:bg-brand/10" aria-label={t('addImage')}>
            <ImageIcon className="h-5 w-5" />
          </button>
          <input ref={fileRef} type="file" accept="image/*" multiple hidden onChange={(e) => addFiles(e.target.files)} />
          <div className="flex items-center gap-3">
            {content.length > 0 && (
              <span className={`text-sm ${tooLong ? 'text-red-500' : 'text-muted'}`}>{remaining}</span>
            )}
            <button type="button" disabled={!canSubmit} onClick={submit} className="btn-primary py-1.5">
              {replyToId ? t('reply') : t('submit')}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
