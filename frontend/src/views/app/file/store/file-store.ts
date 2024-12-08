import { defineStore } from 'pinia';
import { FileItem } from '../types/file-item';

const useFileStore = defineStore('file-manage', {
  state: () => ({
    current: null as FileItem | null,
    tree: [
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
    ] as FileItem[],
    loading: false,
    data: {
      items: [
        {
          path: 'idb-prd/apps/my-sql/aaa',
          name: 'aaa',
          mod_time: '2021-08-01 12:00:00',
          mode: '0755',
          user: 'root',
          group: 'root',
          size: 0,
          is_dir: true,
        },
        {
          path: 'idb-prd/apps/my-sql/aab',
          name: 'aab',
          mod_time: '2021-08-01 12:00:00',
          mode: '0755',
          user: 'root',
          group: 'root',
          is_dir: true,
        },
        {
          path: 'idb-prd/apps/my-sql/aac',
          name: 'aac',
          mod_time: '2021-08-01 12:00:00',
          mode: '0755',
          user: 'root',
          group: 'root',
          size: 43232,
        },
      ] as FileItem[],
      total: 3,
      page: 1,
      page_size: 20,
    },
    showHidden: false,
    selected: [] as any[],
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
    load() {
      const { pwd } = this;
      this.$state.loading = true;
      window.setTimeout(() => {
        this.$state.data = {
          page: 1,
          page_size: 20,
          total: 2,
          items: [
            {
              name: pwd.split('/').pop() + '-1',
              path: pwd + '/' + pwd.split('/').pop() + '-1',
              is_dir: true,
            },
            {
              name: pwd.split('/').pop() + '-2',
              path: pwd + '/' + pwd.split('/').pop() + '-2',
              is_dir: true,
            },
          ] as FileItem[],
        };

        this.$state.loading = false;
      }, 1000);
    },
    loadChildren(treeItem: FileItem) {
      treeItem.loading = true;
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
    handleGoto(path: string) {
      console.log('goto', path);
    },
    handleTreeItemSelect(treeItem: FileItem) {
      if (this.$state.current?.path !== treeItem?.path) {
        this.$state.current = treeItem;
        this.load();
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
        this.loadChildren(treeItem);
      }
    },
    handleBack() {
      if (!this.current || this.pwd === '/') {
        return;
      }
      this.current = this.getParent(this.current);
      this.load();
    },
    handleSelected(selected: any[]) {
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
    handlePaste() {
      console.log('paste', this.selected);
      // todo submit paste
      this.clearSelected();
      this.load();
    },
    handleOpen(item: FileItem) {
      const treeItem = this.getItemByPath(item.path);
      if (!treeItem) {
        return;
      }
      if (treeItem.is_dir) {
        if (!treeItem?.items) {
          this.loadChildren(treeItem);
        } else if (!treeItem.open) {
          treeItem.open = !treeItem.open;
          this.$state.tree = [...this.$state.tree];
        }
      }
      this.handleTreeItemSelect(treeItem || item);
    },
    handleShowHiddenChange(value: boolean) {
      this.$state.showHidden = value;
      this.load();
    },
    handleCreateFolder() {},
    handleCreateFile() {},
  },
});

export default useFileStore;
