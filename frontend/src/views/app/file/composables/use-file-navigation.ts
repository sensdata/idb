import { ref, Ref, computed, unref } from 'vue';
import { Router } from 'vue-router';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { getFileDetailApi } from '@/api/file';
import { createFileRouteWithPagination } from '@/utils/file-route';
import { useLogger } from '@/composables/use-logger';
import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';
import { FileItem } from '@/components/file/file-editor-drawer/types';
import { FileTreeItem } from '../components/file-tree/type';
import useFileStore from '../store/file-store';

// 常量定义
const FILE_SIZE_LIMIT = 1024 * 1024; // 1MB
const API_DOWNLOAD_PATH = '/api/files/{host}/download';

interface FileNavigationParams {
  store: ReturnType<typeof useFileStore>;
  fileEditorDrawerRef: Ref<InstanceType<typeof FileEditorDrawer> | undefined>;
  router: Router;
  currentHostId: Ref<number | undefined>;
  setLoading?: (loading: boolean) => void;
}

interface UseFileNavigationReturn {
  isGotoTriggered: Ref<boolean>;
  isLoading: Ref<boolean>;
  openFileInEditor: (fileOrPath: FileItem | string) => Promise<void>;
  handleGotoWrapper: (path: string) => void;
  handleTreeItemSelect: (record: FileTreeItem) => void;
  handleTreeItemDoubleClick: (record: FileTreeItem) => void;
}

export const useFileNavigation = (
  params: FileNavigationParams
): UseFileNavigationReturn => {
  const { t } = useI18n();
  const { store, fileEditorDrawerRef, router, currentHostId, setLoading } =
    params;
  const { logError } = useLogger('FileNavigation');

  // 响应式状态
  const isGotoTriggered = ref(false);
  const isLoading = ref(false);

  // 计算属性
  const hostId = computed(() => unref(currentHostId));

  // 工具函数：创建路由配置
  const createRouteConfig = (path: string) => {
    return createFileRouteWithPagination(
      path,
      undefined, // 重置分页参数
      hostId.value ? { id: hostId.value } : {}
    );
  };

  // 工具函数：设置加载状态
  const updateLoadingState = (loading: boolean) => {
    isLoading.value = loading;
    if (setLoading) {
      setLoading(loading);
    }
  };

  // 工具函数：提取目录路径
  const extractDirPath = (filePath: string): string => {
    const lastSlashIndex = filePath.lastIndexOf('/');
    return lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';
  };

  // 工具函数：处理文件下载
  const handleFileDownload = (fileDetail: FileItem) => {
    const downloadUrl = `${API_DOWNLOAD_PATH}?source=${encodeURIComponent(
      fileDetail.path
    )}`;
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = fileDetail.name;
    link.style.display = 'none';

    try {
      document.body.appendChild(link);
      link.click();
    } finally {
      document.body.removeChild(link);
    }
  };

  // 工具函数：统一的导航处理
  const navigateToPath = async (
    path: string,
    updateStore = false
  ): Promise<void> => {
    try {
      updateLoadingState(true);

      const routeConfig = createRouteConfig(path);
      await router.push(routeConfig);

      if (updateStore) {
        store.handleGoto(path);
      }
    } catch (error) {
      logError('Navigation failed:', error);
      Message.error(
        t('app.file.list.message.navigationFailed') || 'Navigation failed'
      );
    } finally {
      updateLoadingState(false);
    }
  };

  /**
   * 包装的导航处理函数
   */
  const handleGotoWrapper = (path: string) => {
    isGotoTriggered.value = true;
    navigateToPath(path, true);
  };

  /**
   * 打开或下载文件
   */
  const openFileInEditor = async (
    fileOrPath: FileItem | string
  ): Promise<void> => {
    try {
      const filePath =
        typeof fileOrPath === 'string' ? fileOrPath : fileOrPath.path;
      const dirPath = extractDirPath(filePath);

      const fileDetail = await getFileDetailApi({
        path: filePath,
        expand: false,
      });

      if (!fileDetail) {
        Message.warning(
          t('app.file.list.message.fileNotFound') || 'File not found'
        );
        return;
      }

      // 更新地址栏显示目录路径
      if (store.pwd !== dirPath) {
        const currentGotoState = isGotoTriggered.value;
        isGotoTriggered.value = false;
        store.handleGoto(dirPath);
        isGotoTriggered.value = currentGotoState;
      }

      if (fileDetail.size > FILE_SIZE_LIMIT) {
        // 大文件下载处理
        Message.info(
          t('app.file.list.message.largeFileDownload') ||
            'Large file will be downloaded'
        );
        handleFileDownload(fileDetail);
      } else {
        // 小文件编辑器处理
        const drawer = unref(fileEditorDrawerRef);
        if (drawer) {
          drawer.setFile(fileDetail);
          drawer.show();
        }
      }
    } catch (error) {
      logError('Failed to open file:', error);
      Message.error(
        t('app.file.list.message.fileOpenFailed') || 'Failed to open file'
      );
    }
  };

  /**
   * 文件树选择处理
   */
  const handleTreeItemSelect = (record: FileTreeItem) => {
    if (!record) return;

    if (record.is_dir) {
      navigateToPath(record.path);

      // 更新store状态
      store.$patch({
        current: record,
        selected: [record] as FileItem[],
        addressItems: [],
      });
    } else {
      store.handleTreeItemSelect(record);
    }
  };

  /**
   * 文件树双击处理
   */
  const handleTreeItemDoubleClick = (record: FileTreeItem) => {
    if (record.is_dir) {
      // 目录双击导航
      navigateToPath(record.path);
      store.handleOpen(record as FileItem);
    } else {
      // 文件双击处理
      const parentDir = extractDirPath(record.path);

      // 确保在正确目录
      if (store.pwd !== parentDir) {
        const routeConfig = createRouteConfig(parentDir);
        router.push(routeConfig).catch((error) => {
          logError('Failed to navigate to parent directory:', error);
          Message.error(
            t('app.file.list.message.navigationFailed') || 'Navigation failed'
          );
        });
      }

      // 打开文件
      openFileInEditor(record.path);
    }
  };

  return {
    isGotoTriggered,
    isLoading,
    openFileInEditor,
    handleGotoWrapper,
    handleTreeItemSelect,
    handleTreeItemDoubleClick,
  };
};
