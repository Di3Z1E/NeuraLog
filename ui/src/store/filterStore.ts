import { create } from 'zustand';
import type { LogLevel } from '../types';

interface FilterState {
  selectedNamespace: string | null;
  selectedPod: string | null;
  searchTerm: string;
  level: LogLevel;
  tailMode: boolean;
  liveMode: boolean;
  setSelected: (namespace: string, pod: string) => void;
  setSearch: (term: string) => void;
  setLevel: (level: LogLevel) => void;
  setTailMode: (on: boolean) => void;
  setLiveMode: (on: boolean) => void;
}

export const useFilterStore = create<FilterState>((set) => ({
  selectedNamespace: null,
  selectedPod: null,
  searchTerm: '',
  level: 'ALL',
  tailMode: true,
  liveMode: true,
  setSelected: (selectedNamespace, selectedPod) =>
    set({ selectedNamespace, selectedPod, tailMode: true }),
  setSearch: (searchTerm) => set({ searchTerm }),
  setLevel: (level) => set({ level }),
  setTailMode: (tailMode) => set({ tailMode }),
  setLiveMode: (liveMode) => set({ liveMode }),
}));
