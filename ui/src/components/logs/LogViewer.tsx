import { useEffect, useRef, useCallback } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { LogLine } from './LogLine';
import { useWebSocket } from '../../hooks/useWebSocket';
import { useLogBuffer } from '../../hooks/useLogBuffer';
import { useLogs, toRFC3339 } from '../../api/logs';
import { useFilterStore } from '../../store/filterStore';
import { buildDownloadUrl } from '../../api/client';

export function LogViewer() {
  const {
    selectedNamespace: ns,
    selectedPod: pod,
    searchTerm,
    level,
    liveMode,
    tailMode,
    dateFrom,
    dateTo,
    setTailMode,
  } = useFilterStore();

  const parentRef = useRef<HTMLDivElement>(null);
  const { lines, addLine, setLines, clearLines } = useLogBuffer();

  // Clear buffer on pod change
  useEffect(() => {
    clearLines();
  }, [ns, pod, clearLines]);

  const fromRFC = toRFC3339(dateFrom);
  const toRFC = toRFC3339(dateTo);

  // Historical logs for non-live mode
  const { data: historyLines } = useLogs(ns ?? '', pod ?? '', {
    lines: 2000,
    search: liveMode ? undefined : searchTerm,
    level: liveMode ? undefined : level,
    from: liveMode ? undefined : (fromRFC || undefined),
    to: liveMode ? undefined : (toRFC || undefined),
  });

  useEffect(() => {
    if (!liveMode && historyLines) {
      setLines(historyLines);
    }
  }, [liveMode, historyLines, setLines]);

  const handleMessage = useCallback((line: string) => addLine(line), [addLine]);

  const { status } = useWebSocket(
    liveMode ? ns : null,
    liveMode ? pod : null,
    handleMessage,
  );

  // Client-side filter for live lines (history is filtered server-side)
  const filtered = liveMode
    ? lines.filter((line) => {
        if (searchTerm && !line.toLowerCase().includes(searchTerm.toLowerCase())) return false;
        if (level !== 'ALL' && !line.toUpperCase().includes(level)) return false;
        return true;
      })
    : lines;

  const virtualizer = useVirtualizer({
    count: filtered.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 22,
    overscan: 30,
  });

  useEffect(() => {
    if (tailMode && filtered.length > 0) {
      virtualizer.scrollToIndex(filtered.length - 1, { align: 'end' });
    }
  }, [filtered.length, tailMode, virtualizer]);

  const onScroll = useCallback(() => {
    const el = parentRef.current;
    if (!el) return;
    const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 60;
    setTailMode(nearBottom);
  }, [setTailMode]);

  if (!ns || !pod) {
    return (
      <div className="log-viewer log-viewer--empty">
        <div className="log-viewer__empty-state">
          <div className="log-viewer__empty-icon">⬡</div>
          <div className="log-viewer__empty-title">No pod selected</div>
          <div className="log-viewer__empty-sub">Pick a namespace and pod from the sidebar to start viewing logs</div>
        </div>
      </div>
    );
  }

  const statusLabel =
    status === 'open'
      ? '● LIVE'
      : status === 'connecting'
      ? '◌ CONNECTING'
      : '✕ DISCONNECTED';

  const downloadUrl = buildDownloadUrl(
    ns,
    pod,
    !liveMode && fromRFC ? fromRFC : undefined,
    !liveMode && toRFC ? toRFC : undefined,
  );

  return (
    <div className="log-viewer">
      <div className="log-viewer__header">
        <div className="log-viewer__pod-label">
          <span className="log-viewer__ns">{ns}</span>
          <span className="log-viewer__sep"> / </span>
          <span className="log-viewer__pod">{pod}</span>
        </div>
        <div className="log-viewer__meta">
          {liveMode && (
            <span className={`log-viewer__status log-viewer__status--${status}`}>
              {statusLabel}
            </span>
          )}
          <span className="log-viewer__count">{filtered.length.toLocaleString()} lines</span>
          <a
            className="log-viewer__download"
            href={downloadUrl}
            download
          >
            ↓ Download
          </a>
        </div>
      </div>

      <div ref={parentRef} className="log-viewer__scroll" onScroll={onScroll}>
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative' }}>
          {virtualizer.getVirtualItems().map((item) => (
            <div
              key={item.key}
              style={{
                position: 'absolute',
                top: 0,
                left: 0,
                width: '100%',
                transform: `translateY(${item.start}px)`,
              }}
            >
              <LogLine
                line={filtered[item.index]}
                index={item.index}
                searchTerm={searchTerm}
              />
            </div>
          ))}
        </div>
      </div>

      {!tailMode && filtered.length > 0 && (
        <button
          className="log-viewer__tail-btn"
          onClick={() => {
            setTailMode(true);
            virtualizer.scrollToIndex(filtered.length - 1, { align: 'end' });
          }}
        >
          ↓ Jump to latest
        </button>
      )}
    </div>
  );
}
