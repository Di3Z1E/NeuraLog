export interface PodInfo {
  namespace: string;
  name: string;
  status: 'Running' | 'Pending' | 'Failed' | 'Succeeded' | 'stopped' | string;
  node?: string;
  hasLogs: boolean;
}

export interface LogsResponse {
  namespace: string;
  pod: string;
  lines: string[];
  count: number;
}

export interface PodsResponse {
  pods: PodInfo[];
}

export type LogLevel = 'ALL' | 'TRACE' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL';

export type WsStatus = 'connecting' | 'open' | 'closed' | 'error';
