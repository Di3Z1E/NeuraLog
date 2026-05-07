import { useQuery } from '@tanstack/react-query';
import { fetchJson } from './client';
import type { LogsResponse } from '../types';

interface LogParams {
  lines?: number;
  search?: string;
  level?: string;
}

export function useLogs(namespace: string, pod: string, params: LogParams = {}) {
  return useQuery({
    queryKey: ['logs', namespace, pod, params],
    queryFn: () =>
      fetchJson<LogsResponse>(`/api/v1/logs/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}`, {
        lines: params.lines ?? 2000,
        search: params.search,
        level: params.level !== 'ALL' ? params.level : undefined,
      }),
    enabled: Boolean(namespace && pod),
    staleTime: 5_000,
    select: (data) => data.lines ?? [],
  });
}
