import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchJson } from './client';
import type { AppConfig } from '../types';

export function useConfig() {
  return useQuery<AppConfig>({
    queryKey: ['config'],
    queryFn: () => fetchJson<AppConfig>('/api/v1/config'),
    staleTime: 30_000,
  });
}

export function useUpdateConfig() {
  const qc = useQueryClient();
  return useMutation<AppConfig, Error, AppConfig>({
    mutationFn: (cfg) =>
      fetch('/api/v1/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cfg),
      }).then(async (res) => {
        if (!res.ok) {
          const text = await res.text();
          throw new Error(text || `${res.status} ${res.statusText}`);
        }
        return res.json() as Promise<AppConfig>;
      }),
    onSuccess: (data) => {
      qc.setQueryData(['config'], data);
    },
  });
}
