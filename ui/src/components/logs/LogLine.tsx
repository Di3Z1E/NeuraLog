import React from 'react';

const LEVEL_RE = [
  { re: /\b(FATAL|PANIC)\b/i,   cls: 'fatal' },
  { re: /\b(ERROR|ERR)\b/i,     cls: 'error' },
  { re: /\b(WARN|WARNING)\b/i,  cls: 'warn'  },
  { re: /\b(INFO)\b/i,          cls: 'info'  },
  { re: /\b(DEBUG)\b/i,         cls: 'debug' },
  { re: /\b(TRACE)\b/i,         cls: 'trace' },
];

function detectLevel(line: string): string {
  const head = line.slice(0, 120);
  for (const { re, cls } of LEVEL_RE) {
    if (re.test(head)) return cls;
  }
  return 'default';
}

function renderLine(line: string, searchTerm: string): React.ReactNode {
  // Split on [REDACTED:...] tokens first, then apply search highlight
  const redactedParts = line.split(/(\[REDACTED:[^\]]+\])/g);

  return redactedParts.map((part, i) => {
    if (part.startsWith('[REDACTED:')) {
      return <span key={i} className="log-line__redacted">{part}</span>;
    }
    if (!searchTerm) return part;

    const escaped = searchTerm.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const chunks = part.split(new RegExp(`(${escaped})`, 'gi'));
    return chunks.map((chunk, j) =>
      chunk.toLowerCase() === searchTerm.toLowerCase()
        ? <mark key={j} className="log-line__highlight">{chunk}</mark>
        : chunk,
    );
  });
}

interface Props {
  line: string;
  index: number;
  searchTerm: string;
}

export const LogLine = React.memo(function LogLine({ line, index, searchTerm }: Props) {
  const level = detectLevel(line);
  return (
    <div className={`log-line log-line--${level}`}>
      <span className="log-line__num">{index + 1}</span>
      <span className="log-line__text">{renderLine(line, searchTerm)}</span>
    </div>
  );
});
