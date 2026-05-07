import React from 'react';
import type { PodInfo } from '../../types';

function dotClass(status: string): string {
  switch (status.toLowerCase()) {
    case 'running':   return 'pod-dot--running';
    case 'pending':   return 'pod-dot--pending';
    case 'failed':    return 'pod-dot--failed';
    case 'succeeded': return 'pod-dot--succeeded';
    default:          return 'pod-dot--stopped';
  }
}

interface Props {
  pod: PodInfo;
  selected?: boolean;
  onClick?: () => void;
}

export function PodBadge({ pod, selected, onClick }: Props) {
  return (
    <button
      className={`pod-badge${selected ? ' pod-badge--selected' : ''}`}
      onClick={onClick}
      title={`${pod.namespace}/${pod.name}  •  ${pod.status}`}
    >
      <span className={`pod-dot ${dotClass(pod.status)}`} />
      <span className="pod-badge__name">{pod.name}</span>
      {!pod.hasLogs && pod.status.toLowerCase() !== 'running' && (
        <span className="pod-badge__empty">no logs</span>
      )}
    </button>
  );
}
