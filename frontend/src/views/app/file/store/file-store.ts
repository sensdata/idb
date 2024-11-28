import { defineStore } from 'pinia';

const useFileStore = defineStore('file-manage', {
  state: () => ({
    pwd: 'idb-prd/apps/',
    tree: [
      {
        path: 'idb-prd/apps/my-sql/aaa',
        name: 'aaa',
        is_dir: true,
      },
      {
        path: 'idb-prd/apps/my-sql/aab',
        name: 'aab',
        is_dir: true,
      },
      {
        path: 'idb-prd/apps/my-sql/aac',
        name: 'aac',
      },
    ] as any[],
    data: {
      items: [
        {
          path: 'idb-prd/apps/my-sql/aaa',
          name: 'aaa',
          is_dir: true,
        },
        {
          path: 'idb-prd/apps/my-sql/aab',
          name: 'aab',
          is_dir: true,
        },
        {
          path: 'idb-prd/apps/my-sql/aac',
          name: 'aac',
        },
      ],
      page: 1,
      page_size: 20,
    },
    showHidden: false,
    selected: [] as any[],
    copyActive: false,
    cutActive: false,
  }),

  getters: {
    pasteVisible(state) {
      return (state.copyActive || state.cutActive) && state.selected.length > 0;
    },
  },
  actions: {
    clearSelected() {
      this.$state.selected = [];
      this.$state.copyActive = false;
      this.$state.cutActive = false;
    },
    handleCopy() {
      this.$state.cutActive = false;
      this.$state.copyActive = true;
    },
    handleCut() {
      this.$state.copyActive = false;
      this.$state.cutActive = true;
    },
    handleSelected(selected: any[]) {
      this.$state.selected = selected;
    },
  },
});

export default useFileStore;
