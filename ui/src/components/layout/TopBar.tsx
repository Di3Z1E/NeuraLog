import { useFilterStore } from '../../store/filterStore';
import type { LogLevel } from '../../types';

const LEVELS: LogLevel[] = ['ALL', 'TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'];

export function TopBar() {
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

      <div className="topbar__controls">
        <div className="topbar__search-wrap">
          <svg className="topbar__search-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="6.5" cy="6.5" r="4" />
            <path d="M11 11l3 3" />
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
            <button className="topbar__clear" onClick={() => setSearch('')} aria-label="Clear search">✕</button>
          )}
        </div>

        <select
          className="topbar__level-select"
          value={level}
          onChange={(e) => setLevel(e.target.value as LogLevel)}
        >
          {LEVELS.map((l) => (
            <option key={l} value={l}>{l}</option>
          ))}
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
              <button className="topbar__clear" onClick={clearDateRange} aria-label="Clear date range" title="Clear date range">✕</button>
            )}
          </div>
        )}

        <button
          className={`topbar__live-btn${liveMode ? ' topbar__live-btn--on' : ''}`}
          onClick={() => setLiveMode(!liveMode)}
          title={liveMode ? 'Switch to history mode' : 'Switch to live mode'}
        >
          <span className={`topbar__live-dot${liveMode ? ' topbar__live-dot--on' : ''}`} />
          {liveMode ? 'LIVE' : 'HISTORY'}
        </button>
      </div>
    </header>
  );
}
