import { writable } from 'svelte/store';

export type FinalResult = {
  imageUrl: string;
  summary: string;
  isDemo: boolean;
};

export const finalResult = writable<FinalResult | null>(null);
