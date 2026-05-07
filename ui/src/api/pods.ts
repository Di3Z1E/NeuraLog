import { useQuery } from '@tanstack/react-query';
import { fetchJson } from './client';
import type { PodsResponse, PodInfo } from '../types';

export function usePods(namespace?: string) {
  return useQuery({
    queryKey: ['pods', namespace ?? ''],
    queryFn: () =>
      fetchJson<PodsResponse>('/api/v1/pods', namespace ? { namespace } : undefined),
    refetchInterval: 15_000,
    select: (data): PodInfo[] => data.pods ?? [],
  });
}
