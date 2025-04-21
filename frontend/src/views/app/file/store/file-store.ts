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
    /**
     * 初始化文件树，加载根目录下的所有文件和文件夹
     */
    initTree() {
      searchFileListApi({
        page: 1,
        page_size: 100,
        show_hidden: true,
        path: this.pwd,
      }).then((res) => {
        this.$state.tree = res.items;
      });
    },

    /**
     * 通过路径查找文件树中的项目
     * @param path 文件或文件夹路径
     * @returns 找到的文件树项目，如果未找到则返回null
     */
    getItemByPath(path: string) {
      function findItemByPath(
        tree: FileTreeItem[],
        targetPath: string
      ): FileTreeItem | null {
        for (const node of tree) {
          if (node.path === targetPath) {
            return node;
          }
          // 如果当前节点有子项，则递归查找
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

    /**
     * 获取文件或文件夹的父级文件夹
     * @param item 文件或文件夹项目
     * @returns 父级文件夹项目，如果未找到则返回null
     */
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

          // 如果当前节点的子项中包含目标路径，则当前节点为父节点
          if (node.items) {
            for (const child of node.items) {
              if (child.path === targetPath) {
                return node;
              }
            }
          }

          // 递归查找子树
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

    /**
     * 加载文件夹的子项（包含文件和子文件夹）
     * @param treeItem 需要加载子项的文件夹
     */
    async loadTreeChildren(treeItem: FileTreeItem) {
      // 只有文件夹才能加载子项
      if (!treeItem.is_dir) {
        return;
      }

      treeItem.loading = true;
      const data = await searchFileListApi({
        page: 1,
        page_size: 500,
        show_hidden: true,
        path: treeItem.path,
      });

      treeItem.items = data.items || [];
      treeItem.open = true;
      treeItem.loading = false;
      // 触发视图更新
      this.$state.tree = [...this.$state.tree];
    },

    /**
     * 处理文件树项目选择
     * @param treeItem 被选择的文件或文件夹项目
     */
    handleTreeItemSelect(treeItem: FileTreeItem) {
      if (this.$state.current?.path !== treeItem?.path) {
        this.$state.current = treeItem;
      }
    },

    /**
     * 处理文件树项目的展开/折叠状态变化
     * @param treeItem 文件或文件夹项目
     * @param open 是否展开
     */
    handleTreeItemOpenChange(treeItem: FileTreeItem, open: boolean) {
      // 只有文件夹才能展开/折叠
      if (!treeItem.is_dir) {
        return;
      }

      // 更新树结构以触发视图更新
      const updateTree = () => {
        this.$state.tree = [...this.$state.tree];
      };

      // 处理折叠操作
      if (!open) {
        treeItem.open = false;
        updateTree();
        return;
      }

      // 如果已有子项但未展开，直接展开
      if (treeItem.items?.length) {
        treeItem.open = true;
        updateTree();
        return;
      }

      // 如果正在加载中，不执行任何操作
      if (treeItem.loading) {
        return;
      }

      // 加载子项（默认情况）
      this.loadTreeChildren(treeItem);
    },

    /**
     * 处理地址栏搜索
     * @param payload 搜索参数对象
     */
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

    /**
     * 处理打开文件或文件夹
     * @param item 文件或文件夹项目
     */
    handleOpen(item: FileItem) {
      // 找到树中对应的节点
      const treeItem = this.getItemByPath(item.path);

      // 如果项目不存在于树中
      if (!treeItem) {
        // 只处理目录项目
        if (item.is_dir) {
          this.handleTreeItemSelect(item);
        }
        return;
      }

      // 如果是文件夹，处理展开状态
      if (treeItem.is_dir) {
        // 未加载子项时，加载子项
        if (!treeItem.items) {
          this.loadTreeChildren(treeItem);
        }
        // 已有子项但未展开时，展开子项
        else if (!treeItem.open) {
          treeItem.open = true;
          this.$state.tree = [...this.$state.tree];
        }
        // 其他情况（已展开）则不做操作
      }

      // 无论是文件还是文件夹，都更新当前选中项
      this.handleTreeItemSelect(treeItem);
    },

    /**
     * 处理导航到指定路径
     * @param path 目标路径
     */
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

    /**
     * 处理返回上一级目录
     */
    handleBack() {
      if (!this.current || this.pwd === '/') {
        return;
      }

      // 获取当前路径
      const currentPath = this.pwd;

      // 如果当前路径是根路径，则不执行任何操作
      if (currentPath === '/') {
        return;
      }

      // 移除尾随斜杠（如果有）
      const normalizedPath =
        currentPath.endsWith('/') && currentPath !== '/'
          ? currentPath.slice(0, -1)
          : currentPath;

      // 找到最后一个斜杠的位置
      const lastSlashIndex = normalizedPath.lastIndexOf('/');

      // 如果没有找到斜杠或者斜杠在开头（即路径是根目录下的文件/文件夹），则导航到根目录
      const parentPath =
        lastSlashIndex <= 0 ? '/' : normalizedPath.slice(0, lastSlashIndex);

      // 导航到父级路径
      this.handleGoto(parentPath);
    },

    /**
     * 处理选中项变化
     * @param selected 选中的文件或文件夹项目数组
     */
    handleSelected(selected: FileItem[]) {
      this.$state.selected = selected;
    },

    /**
     * 清除所有选中项
     */
    clearSelected() {
      this.$state.selected = [];
      this.$state.copyActive = false;
      this.$state.cutActive = false;
    },

    /**
     * 处理复制操作
     */
    handleCopy() {
      this.$state.cutActive = false;
      this.$state.copyActive = true;
    },

    /**
     * 处理剪切操作
     */
    handleCut() {
      this.$state.copyActive = false;
      this.$state.cutActive = true;
    },
  },
});

export default useFileStore;
