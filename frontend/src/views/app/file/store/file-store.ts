import { defineStore } from 'pinia';
import { getFileDetailApi, searchFileListApi } from '@/api/file';
import { SimpleFileInfoEntity } from '@/entity/FileInfo';
import { FileItem } from '../types/file-item';
import { FileTreeItem } from '../components/file-tree/type';

const useFileStore = defineStore('file-manage', {
  state: () => ({
    current: null as SimpleFileInfoEntity | null,
    tree: [] as FileTreeItem[],
    addressItems: [] as FileItem[],
    showHidden: false,
    selected: [] as FileItem[],
    copyActive: false,
    cutActive: false,
  }),

  getters: {
    pwd: (state) => state.current?.path || '/',
    pasteVisible(state) {
      return (state.copyActive || state.cutActive) && state.selected.length > 0;
    },
    decompressVisible(state) {
      return (
        state.selected.length === 1 &&
        state.selected[0]?.mime_type === 'application/x-gzip'
      );
    },
  },
  actions: {
    initTree() {
      searchFileListApi({
        page: 1,
        page_size: 100,
        show_hidden: true,
        dir: true,
        path: this.pwd,
      }).then((res) => {
        this.$state.tree = res.items;
      });
    },
    getItemByPath(path: string) {
      function findItemByPath(
        tree: FileTreeItem[],
        targetPath: string
      ): FileTreeItem | null {
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
    getParent(item: FileTreeItem) {
      function findParentByPath(
        tree: FileTreeItem[],
        targetPath: string
      ): FileTreeItem | null {
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
    async loadTreeChildren(treeItem: FileTreeItem) {
      treeItem.loading = true;
      const data = await searchFileListApi({
        page: 1,
        page_size: 500,
        show_hidden: true,
        path: treeItem.path,
        dir: true,
      });
      treeItem.items = data.items || [];
      Object.assign(treeItem, { open: true });
      Object.assign(treeItem, { loading: false });
      this.$state.tree = [...this.$state.tree];
    },
    handleTreeItemSelect(treeItem: FileTreeItem) {
      if (this.$state.current?.path !== treeItem?.path) {
        this.$state.current = treeItem;
      }
    },
    handleTreeItemOpenChange(treeItem: FileTreeItem, open: boolean) {
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
    async handleAddressSearch(payload: { path: string; word?: string }) {
      if (!payload.word) {
        this.$state.addressItems = [];
        return;
      }

      const data = await searchFileListApi({
        page: 1,
        page_size: 100,
        show_hidden: this.$state.showHidden,
        path: payload.path,
        dir: true,
        search: payload.word,
      });
      this.$state.addressItems = data.items || [];
    },
    handleOpen(item: FileItem) {
      const treeItem = this.getItemByPath(item.path);
      if (treeItem?.is_dir) {
        if (!treeItem?.items) {
          this.loadTreeChildren(treeItem);
        } else if (!treeItem.open) {
          treeItem.open = !treeItem.open;
          this.$state.tree = [...this.$state.tree];
        }
      }
      if (treeItem?.is_dir || item?.is_dir) {
        this.handleTreeItemSelect(treeItem || item);
      }
    },
    async handleGoto(path: string) {
      const item = await getFileDetailApi({
        path,
        expand: false,
      });
      if (!item) {
        return;
      }
      if (item.is_dir) {
        this.handleOpen(item);
        this.$state.current = item;
        this.$state.addressItems = [];
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
  },
});

export default useFileStore;
