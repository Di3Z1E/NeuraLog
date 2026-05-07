import { useReducer, useCallback } from 'react';

const MAX_LINES = 10_000;

interface State {
  lines: string[];
}

type Action =
  | { type: 'ADD'; line: string }
  | { type: 'SET'; lines: string[] }
  | { type: 'CLEAR' };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD': {
      const next = [...state.lines, action.line];
      return { lines: next.length > MAX_LINES ? next.slice(-MAX_LINES) : next };
    }
    case 'SET':
      return { lines: action.lines.slice(-MAX_LINES) };
    case 'CLEAR':
      return { lines: [] };
  }
}

export function useLogBuffer() {
  const [state, dispatch] = useReducer(reducer, { lines: [] });
  const addLine = useCallback((line: string) => dispatch({ type: 'ADD', line }), []);
  const setLines = useCallback((lines: string[]) => dispatch({ type: 'SET', lines }), []);
  const clearLines = useCallback(() => dispatch({ type: 'CLEAR' }), []);
  return { lines: state.lines, addLine, setLines, clearLines };
}
