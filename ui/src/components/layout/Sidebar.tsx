import { useMemo, useState } from 'react';
import { usePods } from '../../api/pods';
import { PodBadge } from '../pods/PodBadge';
import { useFilterStore } from '../../store/filterStore';
import type { PodInfo } from '../../types';

function groupByNamespace(pods: PodInfo[]): Map<string, PodInfo[]> {
  const map = new Map<string, PodInfo[]>();
  for (const pod of pods) {
    const list = map.get(pod.namespace) ?? [];
    list.push(pod);
    map.set(pod.namespace, list);
  }
  // Stable sort: namespaces alphabetically, pods within each namespace alphabetically
  const sorted = new Map(
    [...map.entries()]
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([ns, list]) => [ns, [...list].sort((a, b) => a.name.localeCompare(b.name))] as const),
  );
  return sorted;
}

export function Sidebar() {
  const { data: pods = [], isLoading } = usePods();
  const { selectedNamespace, selectedPod, setSelected } = useFilterStore();
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());
  const [filter, setFilter] = useState('');

  // Only show pods that have logs on disk — eliminates flicker from transient pods
  const withLogs = useMemo(() => pods.filter((p) => p.hasLogs), [pods]);

  const visible = useMemo(() => {
    const q = filter.trim().toLowerCase();
    if (!q) return withLogs;
    return withLogs.filter(
      (p) => p.name.toLowerCase().includes(q) || p.namespace.toLowerCase().includes(q),
    );
  }, [withLogs, filter]);

  const grouped = useMemo(() => groupByNamespace(visible), [visible]);

  const toggle = (ns: string) =>
    setCollapsed((prev) => {
      const s = new Set(prev);
      s.has(ns) ? s.delete(ns) : s.add(ns);
      return s;
    });

  return (
    <aside className="sidebar">
      <div className="sidebar__head">
        <div className="sidebar__title-row">
          <span className="sidebar__title">Pods</span>
          {withLogs.length > 0 && (
            <span className="sidebar__count-pill">{withLogs.length}</span>
          )}
          {isLoading && withLogs.length === 0 && (
            <span className="sidebar__count-pill sidebar__count-pill--loading">…</span>
          )}
        </div>
        <div className="sidebar__filter-wrap">
          <svg className="sidebar__filter-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.4">
            <circle cx="6.5" cy="6.5" r="4" />
            <path d="M11 11l3 3" strokeLinecap="round" />
          </svg>
          <input
            className="sidebar__filter"
            type="text"
            placeholder="Filter…"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            spellCheck={false}
          />
          {filter && (
            <button className="sidebar__filter-clear" onClick={() => setFilter('')} aria-label="Clear filter">
              ✕
            </button>
          )}
        </div>
      </div>

      <nav className="sidebar__nav">
        {grouped.size === 0 && (
          <div className="sidebar__empty">
            {isLoading ? 'Connecting…' : filter ? 'No matches' : 'No logs collected yet'}
          </div>
        )}

        {Array.from(grouped.entries()).map(([ns, nsPods]) => {
          const open = !collapsed.has(ns);
          return (
            <div key={ns} className="sidebar__group">
              <button
                className="sidebar__ns-btn"
                onClick={() => toggle(ns)}
                aria-expanded={open}
              >
                <svg
                  className={`sidebar__chevron${open ? ' sidebar__chevron--open' : ''}`}
                  viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5"
                  strokeLinecap="round" strokeLinejoin="round"
                >
                  <path d="M6 4l4 4-4 4" />
                </svg>
                <span className="sidebar__ns-name">{ns}</span>
                <span className="sidebar__ns-badge">{nsPods.length}</span>
              </button>

              {open && (
                <div className="sidebar__pod-list">
                  {nsPods.map((pod) => (
                    <PodBadge
                      key={`${pod.namespace}/${pod.name}`}
                      pod={pod}
                      selected={selectedNamespace === pod.namespace && selectedPod === pod.name}
                      onClick={() => setSelected(pod.namespace, pod.name)}
                    />
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </nav>
    </aside>
  );
}
