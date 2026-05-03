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
  return map;
}

export function Sidebar() {
  const { data: pods = [], isLoading } = usePods();
  const { selectedNamespace, selectedPod, setSelected } = useFilterStore();
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set());

  const grouped = useMemo(() => groupByNamespace(pods), [pods]);

  const toggle = (ns: string) =>
    setCollapsed((prev) => {
      const next = new Set(prev);
      next.has(ns) ? next.delete(ns) : next.add(ns);
      return next;
    });

  return (
    <aside className="sidebar">
      <div className="sidebar__header">
        <span className="sidebar__title">Pods</span>
        <span className="sidebar__count">{pods.length}</span>
      </div>

      {isLoading && pods.length === 0 && (
        <div className="sidebar__loading">
          <span className="sidebar__loading-dot" />
          <span className="sidebar__loading-dot" />
          <span className="sidebar__loading-dot" />
        </div>
      )}

      <nav className="sidebar__nav">
        {Array.from(grouped.entries()).map(([ns, nsPods]) => (
          <div key={ns} className="sidebar__ns-group">
            <button className="sidebar__ns-btn" onClick={() => toggle(ns)}>
              <span className={`sidebar__arrow${collapsed.has(ns) ? '' : ' sidebar__arrow--open'}`}>▶</span>
              <span className="sidebar__ns-name">{ns}</span>
              <span className="sidebar__ns-count">{nsPods.length}</span>
            </button>

            {!collapsed.has(ns) && (
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
        ))}
      </nav>
    </aside>
  );
}
