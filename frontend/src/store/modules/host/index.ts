import { defineStore } from 'pinia';
import { getHostListApi } from '@/api/host';
import { HostEntity } from '@/entity/Host';
import { setApiHostId } from '@/helper/api-helper';
import { HostState } from './types';

const useHostStore = defineStore('host', {
  state: (): HostState => ({
    current: undefined,
    currentId: undefined,
    items: [],
    isReady: false,
  }),

  getters: {
    default(): HostEntity {
      return this.items.find((item) => item.is_default) || this.items[0];
    },
  },

  actions: {
    async init() {
      if (!this.isReady) {
        await this.load();
        this.isReady = true;
      }
    },
    async load() {
      const data = await getHostListApi({
        page: 1,
        page_size: 1000,
      });
      this.setItems(data.items);
    },
    async reload() {
      return this.load();
    },
    setItems(items: HostState['items']) {
      this.items = items;
      if (!this.current && items.length) {
        this.setCurrentId(items[0]?.id);
      }
      if (this.currentId && !this.current) {
        this.setCurrentId(this.currentId);
      }
    },
    setCurrentId(hostId?: number) {
      this.currentId = hostId;
      this.current = this.items.find((item) => item.id === hostId);
      setApiHostId(hostId);
    },
  },
});

export default useHostStore;
