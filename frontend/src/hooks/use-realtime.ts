'use client';

import { useEffect, useRef } from 'react';
import { api } from '@/lib/api';

export interface RealtimeEvent {
  type: 'notification' | 'message';
  data: unknown;
}

// useRealtime opens a WebSocket to the API and invokes the handler on each event.
// It transparently reconnects with a small backoff.
export function useRealtime(enabled: boolean, onEvent: (ev: RealtimeEvent) => void) {
  const handlerRef = useRef(onEvent);
  handlerRef.current = onEvent;

  useEffect(() => {
    if (!enabled) return;
    let ws: WebSocket | null = null;
    let closed = false;
    let retry = 0;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const wsURL = () => {
      const base = api.base();
      const u = new URL(base);
      u.protocol = u.protocol === 'https:' ? 'wss:' : 'ws:';
      u.pathname = '/api/ws';
      return u.toString();
    };

    const connect = () => {
      if (closed) return;
      ws = new WebSocket(wsURL());
      ws.onopen = () => {
        retry = 0;
      };
      ws.onmessage = (e) => {
        try {
          handlerRef.current(JSON.parse(e.data) as RealtimeEvent);
        } catch {
          /* ignore malformed frames */
        }
      };
      ws.onclose = () => {
        if (closed) return;
        retry = Math.min(retry + 1, 6);
        timer = setTimeout(connect, 1000 * retry);
      };
      ws.onerror = () => ws?.close();
    };

    connect();

    return () => {
      closed = true;
      if (timer) clearTimeout(timer);
      ws?.close();
    };
  }, [enabled]);
}
