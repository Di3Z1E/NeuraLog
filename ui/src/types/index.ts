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

export interface RedactPattern {
  id: string;
  pattern: string;
  replace: string;
}

export interface AppConfig {
  storageQuotaGB: number;
  rotationMaxMB: number;
  rotationKeepFiles: number;
  retentionDays: number;
  excludeNamespaces: string[];
  redactEnabled: boolean;
  customPatterns: RedactPattern[];
  // read-only, returned by GET but ignored on PUT
  storageUsedGB?: number;
}
