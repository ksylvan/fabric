import { writable } from 'svelte/store';

function createDrawerStore() {
  const { subscribe, set, update } = writable(false);

  return {
    subscribe,
    open: () => set(true),
    close: () => set(false),
    toggle: () => update(v => !v),
  };
}

export const drawerStore = createDrawerStore();
