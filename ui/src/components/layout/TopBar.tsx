import { useFilterStore } from '../../store/filterStore';
import type { LogLevel } from '../../types';

const LEVELS: LogLevel[] = ['ALL', 'TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];

interface Props {
  onSettings: () => void;
}

export function TopBar({ onSettings }: Props) {
  const {
    searchTerm, level, liveMode,
    dateFrom, dateTo,
    setSearch, setLevel, setLiveMode,
    setDateFrom, setDateTo, clearDateRange,
  } = useFilterStore();

  const hasDateRange = Boolean(dateFrom || dateTo);

  return (
    <header className="topbar">
      <div className="topbar__brand">
        <svg className="topbar__logo" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9 3H5a2 2 0 00-2 2v4m6-6h10a2 2 0 012 2v4M9 3v18m0 0h10a2 2 0 002-2V9M9 21H5a2 2 0 01-2-2V9m0 0h18" />
        </svg>
        <span className="topbar__wordmark">NeuraLog</span>
      </div>

      <div className="topbar__center">
        <div className="topbar__search-wrap">
          <svg className="topbar__search-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="6.5" cy="6.5" r="4" />
            <path d="M11 11l3 3" strokeLinecap="round" />
          </svg>
          <input
            className="topbar__search"
            type="text"
            placeholder="Search logs…"
            value={searchTerm}
            onChange={(e) => setSearch(e.target.value)}
            spellCheck={false}
          />
          {searchTerm && (
            <button className="topbar__search-clear" onClick={() => setSearch('')} aria-label="Clear">✕</button>
          )}
        </div>

        <select
          className="topbar__level"
          value={level}
          onChange={(e) => setLevel(e.target.value as LogLevel)}
          aria-label="Log level"
        >
          {LEVELS.map((l) => <option key={l} value={l}>{l}</option>)}
        </select>

        {!liveMode && (
          <div className="topbar__daterange">
            <input
              className="topbar__date"
              type="datetime-local"
              title="From"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
            />
            <span className="topbar__date-sep">→</span>
            <input
              className="topbar__date"
              type="datetime-local"
              title="To"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
            />
            {hasDateRange && (
              <button className="topbar__search-clear" onClick={clearDateRange} title="Clear range">✕</button>
            )}
          </div>
        )}
      </div>

      <div className="topbar__right">
        <div className="topbar__mode-pill" role="group" aria-label="View mode">
          <button
            className={`topbar__mode-btn${liveMode ? ' topbar__mode-btn--active' : ''}`}
            onClick={() => setLiveMode(true)}
          >
            <span className={`topbar__live-dot${liveMode ? ' topbar__live-dot--on' : ''}`} />
            Live
          </button>
          <button
            className={`topbar__mode-btn${!liveMode ? ' topbar__mode-btn--active' : ''}`}
            onClick={() => setLiveMode(false)}
          >
            History
          </button>
        </div>

        <button
          className="topbar__icon-btn"
          onClick={onSettings}
          title="Settings"
          aria-label="Open settings"
        >
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12.22 2h-.44a2 2 0 00-2 2v.18a2 2 0 01-1 1.73l-.43.25a2 2 0 01-2 0l-.15-.08a2 2 0 00-2.73.73l-.22.38a2 2 0 00.73 2.73l.15.1a2 2 0 011 1.72v.51a2 2 0 01-1 1.74l-.15.09a2 2 0 00-.73 2.73l.22.38a2 2 0 002.73.73l.15-.08a2 2 0 012 0l.43.25a2 2 0 011 1.73V20a2 2 0 002 2h.44a2 2 0 002-2v-.18a2 2 0 011-1.73l.43-.25a2 2 0 012 0l.15.08a2 2 0 002.73-.73l.22-.39a2 2 0 00-.73-2.73l-.15-.08a2 2 0 01-1-1.74v-.5a2 2 0 011-1.74l.15-.09a2 2 0 00.73-2.73l-.22-.38a2 2 0 00-2.73-.73l-.15.08a2 2 0 01-2 0l-.43-.25a2 2 0 01-1-1.73V4a2 2 0 00-2-2z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        </button>
      </div>
    </header>
  );
}
