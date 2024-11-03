import { defineStore } from 'pinia';
import { getHostOptionsApi } from '@/api/host';
import { HostState } from './types';

const useHostStore = defineStore('host', {
  state: (): HostState => ({
    current: undefined,
    currentId: undefined,
    items: [],
  }),

  actions: {
    async init() {
      const data = await getHostOptionsApi();
      this.setItems(data.items);
    },
    setItems(items: HostState['items']) {
      this.items = items;
      if (!this.current && items.length) {
        this.setCurrentId(items[0]?.id);
      }
    },
    setCurrentId(hostId: number) {
      this.currentId = hostId;
      this.current = this.items.find((item) => item.id === hostId);
    },
  },
});

export default useHostStore;
