import { defineStore } from 'pinia';
import { getFileInfoApi, getFileListApi, moveFileApi } from '@/api/file';
import { FileItem } from '../types/file-item';

const useFileStore = defineStore('file-manage', {
  state: () => ({
    current: null as FileItem | null,
    tree: [] as FileItem[],
    addressItems: [] as FileItem[],
    selected: [] as FileItem[],
    copyActive: false,
    cutActive: false,
  }),

  getters: {
    pwd: (state) => state.current?.path || '/',
    pasteVisible(state) {
      return (state.copyActive || state.cutActive) && state.selected.length > 0;
    },
  },
  actions: {
    initTree() {
      // getFileListApi().then((res) => {
      //   this.$state.tree = res.items;
      // });
      window.setTimeout(() => {
        this.$state.tree = [
          {
            path: 'idb-prd/apps/my-sql/aaa',
            name: 'aaa',
            is_dir: true,
            loading: false,
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
        ] as FileItem[];
      }, 1000);
    },
    getItemByPath(path: string) {
      function findItemByPath(
        tree: FileItem[],
        targetPath: string
      ): FileItem | null {
        for (const node of tree) {
          if (node.path === targetPath) {
            return node;
          }
          if (node.items) {
            const result = findItemByPath(node.items, targetPath);
            if (result) {
              return result;
            }
          }
        }
        return null;
      }
      return findItemByPath(this.$state.tree, path);
    },
    getParent(item: FileItem) {
      function findParentByPath(
        tree: FileItem[],
        targetPath: string
      ): FileItem | null {
        for (const node of tree) {
          // 跳过不匹配的前缀项，优化性能
          if (!targetPath.startsWith(node.path)) {
            continue;
          }

          if (node.items) {
            for (const child of node.items) {
              if (child.path === targetPath) {
                return node;
              }
            }
          }

          if (node.items) {
            const result = findParentByPath(node.items, targetPath);
            if (result) {
              return result;
            }
          }
        }
        return null;
      }
      return findParentByPath(this.$state.tree, item.path);
    },
    async loadTreeChildren(treeItem: FileItem) {
      treeItem.loading = true;
      // const data = await getFileListApi();
      // treeItem.items = data.items;
      // Object.assign(treeItem, { open: true });
      // Object.assign(treeItem, { loading: false });
      // this.$state.tree = [...this.$state.tree];

      window.setTimeout(() => {
        Object.assign(treeItem, {
          items: [
            {
              name: treeItem.name + '-1',
              path: treeItem.path + '/' + treeItem.name + '-1',
              is_dir: true,
            },
            {
              name: treeItem.name + '-2',
              path: treeItem.path + '/' + treeItem.name + '-2',
              is_dir: true,
            },
          ],
        });
        Object.assign(treeItem, { open: true });
        Object.assign(treeItem, { loading: false });
        this.$state.tree = [...this.$state.tree];
      }, 1000);
    },
    handleTreeItemSelect(treeItem: FileItem) {
      if (this.$state.current?.path !== treeItem?.path) {
        this.$state.current = treeItem;
      }
    },
    handleTreeItemOpenChange(treeItem: FileItem, open: boolean) {
      if (!treeItem.is_dir) {
        return;
      }
      if (!open) {
        Object.assign(treeItem, { open: false });
        this.$state.tree = [...this.$state.tree];
        return;
      }

      if (treeItem.items && !treeItem.open) {
        Object.assign(treeItem, { open: true });
        this.$state.tree = [...this.$state.tree];
        return;
      }

      if (!treeItem.loading && !treeItem.items) {
        this.loadTreeChildren(treeItem);
      }
    },
    handleAddressSearch(payload: { path: string; word?: string }) {
      if (!payload.word) {
        this.$state.addressItems = [];
        return;
      }
      window.setTimeout(() => {
        this.$state.addressItems = [
          {
            name: payload.word + '-1',
            path: payload.path + '/' + payload.word + '-1',
            is_dir: true,
          },
          {
            name: payload.word + '-2',
            path: payload.path + '/' + payload.word + '-2',
            is_dir: true,
          },
        ] as any[];
      }, 1000);
    },
    handleOpen(item: FileItem) {
      const treeItem = this.getItemByPath(item.path);
      if (!treeItem) {
        return;
      }
      if (treeItem.is_dir) {
        if (!treeItem?.items) {
          this.loadTreeChildren(treeItem);
        } else if (!treeItem.open) {
          treeItem.open = !treeItem.open;
          this.$state.tree = [...this.$state.tree];
        }
      }
      this.handleTreeItemSelect(treeItem || item);
    },
    async handleGoto(path: string) {
      const item = await getFileInfoApi({ path });
      if (item?.is_dir) {
        this.handleOpen(item);
        this.$state.current = item;
      }
    },
    handleBack() {
      if (!this.current || this.pwd === '/') {
        return;
      }
      this.current = this.getParent(this.current);
    },
    handleSelected(selected: FileItem[]) {
      this.$state.selected = selected;
    },
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
    async handlePaste() {
      await moveFileApi({
        sources: this.$state.selected.map((item) => item.path),
        dest: this.pwd,
        cover: false,
        type: this.$state.cutActive ? 'move' : 'copy',
      });
      this.clearSelected();
      return true;
    },
  },
});

export default useFileStore;
