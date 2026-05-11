// All paths are relative — nginx (prod) and Vite proxy (dev) handle routing transparently.

export async function fetchJson<T>(
  path: string,
  params?: Record<string, string | number | undefined>,
): Promise<T> {
  const url = new URL(path, window.location.href);
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      if (v !== undefined && v !== '') {
        url.searchParams.set(k, String(v));
      }
    }
  }
  const res = await fetch(url.toString());
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText} — ${path}`);
  }
  return res.json() as Promise<T>;
}

export function buildWsUrl(namespace: string, pod: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  return `${proto}://${window.location.host}/ws?namespace=${encodeURIComponent(namespace)}&pod=${encodeURIComponent(pod)}`;
}

export function buildDownloadUrl(
  namespace: string,
  pod: string,
  from?: string,
  to?: string,
): string {
  const url = `/api/v1/download/${encodeURIComponent(namespace)}/${encodeURIComponent(pod)}`;
  const p = new URLSearchParams();
  if (from) p.set('from', from);
  if (to) p.set('to', to);
  const qs = p.toString();
  return qs ? url + '?' + qs : url;
}
