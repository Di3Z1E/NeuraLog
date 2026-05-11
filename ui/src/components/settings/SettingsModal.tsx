import { useState, useEffect, useCallback } from 'react';
import { useConfig, useUpdateConfig } from '../../api/config';
import type { AppConfig, RedactPattern } from '../../types';

type Tab = 'storage' | 'collection' | 'redaction';

interface Props {
  onClose: () => void;
}

function GbBar({ used, quota }: { used: number; quota: number }) {
  if (quota <= 0) return null;
  const pct = Math.min((used / quota) * 100, 100);
  const color = pct > 90 ? 'var(--lv-error)' : pct > 70 ? 'var(--lv-warn)' : 'var(--accent)';
  return (
    <div className="cfg-quota-bar">
      <div className="cfg-quota-bar__track">
        <div className="cfg-quota-bar__fill" style={{ width: `${pct}%`, background: color }} />
      </div>
      <span className="cfg-quota-bar__label">
        {used.toFixed(2)} GB / {quota} GB used
      </span>
    </div>
  );
}

export function SettingsModal({ onClose }: Props) {
  const { data: remote, isLoading } = useConfig();
  const { mutate: save, isPending, isSuccess, error, reset } = useUpdateConfig();

  const [cfg, setCfg] = useState<AppConfig | null>(null);
  const [tab, setTab] = useState<Tab>('storage');
  const [newNs, setNewNs] = useState('');
  const [newPattern, setNewPattern] = useState('');
  const [newReplace, setNewReplace] = useState('');

  useEffect(() => {
    if (remote) setCfg(remote);
  }, [remote]);

  const patch = useCallback((p: Partial<AppConfig>) => {
    setCfg((prev) => (prev ? { ...prev, ...p } : prev));
    reset();
  }, [reset]);

  const addNs = () => {
    const ns = newNs.trim();
    if (!ns || !cfg || cfg.excludeNamespaces.includes(ns)) return;
    patch({ excludeNamespaces: [...cfg.excludeNamespaces, ns] });
    setNewNs('');
  };

  const removeNs = (ns: string) => {
    if (!cfg) return;
    patch({ excludeNamespaces: cfg.excludeNamespaces.filter((n) => n !== ns) });
  };

  const addPattern = () => {
    const p = newPattern.trim();
    const r = newReplace.trim();
    if (!p || !r || !cfg) return;
    const entry: RedactPattern = { id: crypto.randomUUID(), pattern: p, replace: r };
    patch({ customPatterns: [...cfg.customPatterns, entry] });
    setNewPattern('');
    setNewReplace('');
  };

  const removePattern = (id: string) => {
    if (!cfg) return;
    patch({ customPatterns: cfg.customPatterns.filter((p) => p.id !== id) });
  };

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) onClose();
  };

  if (isLoading || !cfg) {
    return (
      <div className="cfg-overlay" onClick={handleOverlayClick}>
        <div className="cfg-modal cfg-modal--loading">
          <div className="cfg-spinner" />
        </div>
      </div>
    );
  }

  return (
    <div className="cfg-overlay" onClick={handleOverlayClick}>
      <div className="cfg-modal">
        {/* Header */}
        <div className="cfg-header">
          <div className="cfg-header__title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12.22 2h-.44a2 2 0 00-2 2v.18a2 2 0 01-1 1.73l-.43.25a2 2 0 01-2 0l-.15-.08a2 2 0 00-2.73.73l-.22.38a2 2 0 00.73 2.73l.15.1a2 2 0 011 1.72v.51a2 2 0 01-1 1.74l-.15.09a2 2 0 00-.73 2.73l.22.38a2 2 0 002.73.73l.15-.08a2 2 0 012 0l.43.25a2 2 0 011 1.73V20a2 2 0 002 2h.44a2 2 0 002-2v-.18a2 2 0 011-1.73l.43-.25a2 2 0 012 0l.15.08a2 2 0 002.73-.73l.22-.39a2 2 0 00-.73-2.73l-.15-.08a2 2 0 01-1-1.74v-.5a2 2 0 011-1.74l.15-.09a2 2 0 00.73-2.73l-.22-.38a2 2 0 00-2.73-.73l-.15.08a2 2 0 01-2 0l-.43-.25a2 2 0 01-1-1.73V4a2 2 0 00-2-2z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
            Settings
          </div>
          <button className="cfg-header__close" onClick={onClose} aria-label="Close">✕</button>
        </div>

        {/* Tabs */}
        <div className="cfg-tabs">
          {(['storage', 'collection', 'redaction'] as Tab[]).map((t) => (
            <button
              key={t}
              className={`cfg-tab${tab === t ? ' cfg-tab--active' : ''}`}
              onClick={() => setTab(t)}
            >
              {t === 'storage' ? 'Storage' : t === 'collection' ? 'Collection' : 'Redaction'}
            </button>
          ))}
        </div>

        {/* Body */}
        <div className="cfg-body">

          {tab === 'storage' && (
            <div className="cfg-section">
              {cfg.storageUsedGB !== undefined && (
                <GbBar used={cfg.storageUsedGB} quota={cfg.storageQuotaGB} />
              )}

              <div className="cfg-field">
                <label className="cfg-label">Storage quota <span className="cfg-unit">GiB</span></label>
                <div className="cfg-hint">Maximum disk space for all logs. Set to 0 for unlimited.</div>
                <input
                  className="cfg-input"
                  type="number" min="0" step="1"
                  value={cfg.storageQuotaGB}
                  onChange={(e) => patch({ storageQuotaGB: parseFloat(e.target.value) || 0 })}
                />
              </div>

              <div className="cfg-field">
                <label className="cfg-label">Rotation size <span className="cfg-unit">MB per pod</span></label>
                <div className="cfg-hint">Rotate a pod's log file when it reaches this size. Set to 0 to disable rotation.</div>
                <input
                  className="cfg-input"
                  type="number" min="0" step="10"
                  value={cfg.rotationMaxMB}
                  onChange={(e) => patch({ rotationMaxMB: parseInt(e.target.value) || 0 })}
                />
              </div>

              <div className="cfg-field">
                <label className="cfg-label">Rotated files to keep <span className="cfg-unit">per pod</span></label>
                <div className="cfg-hint">Number of old log files to keep after rotation (pod.log.1 … pod.log.N).</div>
                <input
                  className="cfg-input"
                  type="number" min="1" max="20"
                  value={cfg.rotationKeepFiles}
                  onChange={(e) => patch({ rotationKeepFiles: parseInt(e.target.value) || 1 })}
                />
              </div>

              <div className="cfg-field">
                <label className="cfg-label">Retention <span className="cfg-unit">days</span></label>
                <div className="cfg-hint">Log files older than this are deleted by the nightly janitor.</div>
                <input
                  className="cfg-input"
                  type="number" min="1" max="365"
                  value={cfg.retentionDays}
                  onChange={(e) => patch({ retentionDays: parseInt(e.target.value) || 7 })}
                />
              </div>
            </div>
          )}

          {tab === 'collection' && (
            <div className="cfg-section">
              <div className="cfg-field">
                <label className="cfg-label">Excluded namespaces</label>
                <div className="cfg-hint">
                  Logs from these namespaces will not be collected. Changes take effect immediately.
                </div>
                <div className="cfg-tags">
                  {cfg.excludeNamespaces.length === 0 && (
                    <span className="cfg-tags__empty">No exclusions — collecting all namespaces</span>
                  )}
                  {cfg.excludeNamespaces.map((ns) => (
                    <span key={ns} className="cfg-tag">
                      {ns}
                      <button className="cfg-tag__rm" onClick={() => removeNs(ns)} aria-label={`Remove ${ns}`}>✕</button>
                    </span>
                  ))}
                </div>
                <div className="cfg-add-row">
                  <input
                    className="cfg-input cfg-input--grow"
                    type="text"
                    placeholder="namespace name"
                    value={newNs}
                    onChange={(e) => setNewNs(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && addNs()}
                  />
                  <button className="cfg-btn cfg-btn--add" onClick={addNs}>Add</button>
                </div>
              </div>
            </div>
          )}

          {tab === 'redaction' && (
            <div className="cfg-section">
              <div className="cfg-field">
                <label className="cfg-toggle">
                  <div
                    className={`cfg-toggle__track${cfg.redactEnabled ? ' cfg-toggle__track--on' : ''}`}
                    onClick={() => patch({ redactEnabled: !cfg.redactEnabled })}
                    role="checkbox"
                    aria-checked={cfg.redactEnabled}
                    tabIndex={0}
                    onKeyDown={(e) => e.key === ' ' && patch({ redactEnabled: !cfg.redactEnabled })}
                  >
                    <div className="cfg-toggle__thumb" />
                  </div>
                  <span className="cfg-toggle__label">
                    Redaction {cfg.redactEnabled ? 'enabled' : 'disabled'}
                  </span>
                </label>
                <div className="cfg-hint">
                  Strips JWTs, API keys, passwords, DB URLs, credit cards, and custom patterns before writing to disk or streaming.
                  Takes effect immediately without restart.
                </div>
              </div>

              <div className="cfg-field">
                <label className="cfg-label">Custom patterns</label>
                <div className="cfg-hint">Additional regex rules applied on top of the built-in patterns.</div>

                {cfg.customPatterns.length > 0 && (
                  <div className="cfg-pattern-list">
                    {cfg.customPatterns.map((p) => (
                      <div key={p.id} className="cfg-pattern-row">
                        <code className="cfg-pattern-cell cfg-pattern-cell--pattern">{p.pattern}</code>
                        <code className="cfg-pattern-cell cfg-pattern-cell--replace">{p.replace}</code>
                        <button className="cfg-tag__rm" onClick={() => removePattern(p.id)} aria-label="Remove pattern">✕</button>
                      </div>
                    ))}
                  </div>
                )}

                <div className="cfg-add-row cfg-add-row--pattern">
                  <input
                    className="cfg-input cfg-input--grow"
                    type="text"
                    placeholder="regex pattern"
                    value={newPattern}
                    onChange={(e) => setNewPattern(e.target.value)}
                  />
                  <input
                    className="cfg-input cfg-input--grow"
                    type="text"
                    placeholder="[REDACTED:TYPE]"
                    value={newReplace}
                    onChange={(e) => setNewReplace(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && addPattern()}
                  />
                  <button className="cfg-btn cfg-btn--add" onClick={addPattern}>Add</button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="cfg-footer">
          <div className="cfg-footer__status">
            {isSuccess && <span className="cfg-status cfg-status--ok">Saved successfully</span>}
            {error && <span className="cfg-status cfg-status--err">{error.message}</span>}
          </div>
          <div className="cfg-footer__actions">
            <button className="cfg-btn cfg-btn--cancel" onClick={onClose}>Cancel</button>
            <button
              className="cfg-btn cfg-btn--save"
              onClick={() => save(cfg)}
              disabled={isPending}
            >
              {isPending ? 'Saving…' : 'Save changes'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
