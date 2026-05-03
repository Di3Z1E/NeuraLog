import { create } from 'zustand';
import type { LogLevel } from '../types';

interface FilterState {
  selectedNamespace: string | null;
  selectedPod: string | null;
  searchTerm: string;
  level: LogLevel;
  tailMode: boolean;
  liveMode: boolean;
  dateFrom: string;
  dateTo: string;
  setSelected: (namespace: string, pod: string) => void;
  setSearch: (term: string) => void;
  setLevel: (level: LogLevel) => void;
  setTailMode: (on: boolean) => void;
  setLiveMode: (on: boolean) => void;
  setDateFrom: (v: string) => void;
  setDateTo: (v: string) => void;
  clearDateRange: () => void;
}

export const useFilterStore = create<FilterState>((set) => ({
  selectedNamespace: null,
  selectedPod: null,
  searchTerm: '',
  level: 'ALL',
  tailMode: true,
  liveMode: true,
  dateFrom: '',
  dateTo: '',
  setSelected: (selectedNamespace, selectedPod) =>
    set({ selectedNamespace, selectedPod, tailMode: true }),
  setSearch: (searchTerm) => set({ searchTerm }),
  setLevel: (level) => set({ level }),
  setTailMode: (tailMode) => set({ tailMode }),
  setLiveMode: (liveMode) => set({ liveMode }),
  setDateFrom: (dateFrom) => set({ dateFrom }),
  setDateTo: (dateTo) => set({ dateTo }),
  clearDateRange: () => set({ dateFrom: '', dateTo: '' }),
}));
