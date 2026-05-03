import { useQuery } from '@tanstack/react-query';
import { fetchJson } from './client';
import type { LogsResponse } from '../types';

interface LogParams {
  lines?: number;
  search?: string;
  level?: string;
  from?: string;
  to?: string;
}

// Convert a datetime-local value ("YYYY-MM-DDTHH:mm") to RFC3339 UTC.
export function toRFC3339(v: string): string {
  if (!v) return '';
  // datetime-local has no timezone — treat as UTC
  return v.length === 16 ? v + ':00Z' : v;
}

export function useLogs(namespace: string, pod: string, params: LogParams = {}) {
  const hasDateRange = Boolean(params.from || params.to);
  return useQuery({
    queryKey: ['logs', namespace, pod, params],
    queryFn: () =>
      fetchJson<LogsResponse>(`/api/v1/logs/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}`, {
        // Omit lines cap when date range is active — backend returns all matches.
        lines: hasDateRange ? undefined : (params.lines ?? 2000),
        search: params.search,
        level: params.level !== 'ALL' ? params.level : undefined,
        from: params.from,
        to: params.to,
      }),
    enabled: Boolean(namespace && pod),
    staleTime: 5_000,
    select: (data) => data.lines ?? [],
  });
}
