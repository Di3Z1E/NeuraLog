import type { PodInfo } from '../../types';

type StatusKey = 'running' | 'pending' | 'failed' | 'succeeded' | 'stopped';

function statusKey(s: string): StatusKey {
  switch (s.toLowerCase()) {
    case 'running':   return 'running';
    case 'pending':   return 'pending';
    case 'failed':    return 'failed';
    case 'succeeded': return 'succeeded';
    default:          return 'stopped';
  }
}

interface Props {
  pod: PodInfo;
  selected?: boolean;
  onClick?: () => void;
}

export function PodBadge({ pod, selected, onClick }: Props) {
  const sk = statusKey(pod.status);
  return (
    <button
      className={`pod-badge pod-badge--${sk}${selected ? ' pod-badge--selected' : ''}`}
      onClick={onClick}
      title={`${pod.namespace}/${pod.name} · ${pod.status}${pod.node ? ` · ${pod.node}` : ''}`}
    >
      <span className={`pod-dot pod-dot--${sk}`} />
      <span className="pod-badge__name">{pod.name}</span>
    </button>
  );
}
