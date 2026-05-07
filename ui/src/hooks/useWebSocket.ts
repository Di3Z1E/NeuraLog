import { useEffect, useRef, useState } from 'react';
import { buildWsUrl } from '../api/client';
import type { WsStatus } from '../types';

export function useWebSocket(
  namespace: string | null,
  pod: string | null,
  onMessage: (line: string) => void,
): { status: WsStatus } {
  const [status, setStatus] = useState<WsStatus>('closed');

  // Keep onMessage in a ref to avoid reconnecting when parent re-renders
  const onMessageRef = useRef(onMessage);
  useEffect(() => {
    onMessageRef.current = onMessage;
  });

  useEffect(() => {
    if (!namespace || !pod) {
      setStatus('closed');
      return;
    }

    let ws: WebSocket | null = null;
    let retryDelay = 1_000;
    let retryTimeout: ReturnType<typeof setTimeout> | null = null;
    let cancelled = false;

    function connect() {
      if (cancelled) return;
      setStatus('connecting');
      ws = new WebSocket(buildWsUrl(namespace!, pod!));

      ws.onopen = () => {
        if (cancelled) { ws?.close(); return; }
        setStatus('open');
        retryDelay = 1_000;
      };

      ws.onmessage = (e: MessageEvent<string>) => {
        onMessageRef.current(e.data);
      };

      ws.onerror = () => {
        setStatus('error');
      };

      ws.onclose = () => {
        if (cancelled) return;
        setStatus('closed');
        retryTimeout = setTimeout(() => {
          retryDelay = Math.min(retryDelay * 2, 30_000);
          connect();
        }, retryDelay);
      };
    }

    connect();

    return () => {
      cancelled = true;
      if (retryTimeout) clearTimeout(retryTimeout);
      ws?.close();
      setStatus('closed');
    };
  }, [namespace, pod]);

  return { status };
}
