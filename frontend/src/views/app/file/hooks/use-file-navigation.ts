import { ref, Ref } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { getFileDetailApi } from '@/api/file';
import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';
import { FileItem } from '@/components/file/file-editor-drawer/types';
import { FileTreeItem } from '../components/file-tree/type';
import useFileStore from '../store/file-store';

interface FileNavigationParams {
  store: ReturnType<typeof useFileStore>;
  fileEditorDrawerRef: Ref<InstanceType<typeof FileEditorDrawer> | undefined>;
}

export const useFileNavigation = (params: FileNavigationParams) => {
  const { t } = useI18n();
  const { store, fileEditorDrawerRef } = params;

  // Flag to track if selection was triggered by goto
  const isGotoTriggered = ref(false);

  // Create a wrapper for handleGoto to track the source of selection
  const handleGotoWrapper = (path: string) => {
    isGotoTriggered.value = true;
    store.handleGoto(path);
  };

  /**
   * 打开或下载文件
   */
  const openFileInEditor = async (fileOrPath: FileItem | string) => {
    try {
      const filePath =
        typeof fileOrPath === 'string' ? fileOrPath : fileOrPath.path;

      // 提取目录路径
      const lastSlashIndex = filePath.lastIndexOf('/');
      const dirPath =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      const fileDetail = await getFileDetailApi({
        path: filePath,
        expand: false,
      });

      if (!fileDetail) return;

      // 更新地址栏显示目录路径
      if (store.pwd !== dirPath) {
        // 使用handleGoto方法但跳过触发isGotoTriggered标志
        // 我们需要临时禁用该标志，因为我们已经在goto操作中
        const currentGotoState = isGotoTriggered.value;
        isGotoTriggered.value = false;
        store.handleGoto(dirPath);
        isGotoTriggered.value = currentGotoState;
      }

      if (fileDetail.size > 1048576) {
        // 大文件以下载形式处理
        Message.info(t('app.file.list.message.largeFileDownload'));

        // 创建下载链接并触发下载
        const downloadUrl = `/api/files/{host}/download?source=${encodeURIComponent(
          fileDetail.path
        )}`;
        const a = document.createElement('a');
        a.href = downloadUrl;
        a.download = fileDetail.name;
        a.click();
      } else {
        // 小文件打开编辑器
        fileEditorDrawerRef.value?.setFile(fileDetail);
        fileEditorDrawerRef.value?.show();
      }
    } catch (error) {
      console.error('获取文件详情失败:', error);
      Message.error(t('app.file.list.message.fileOpenFailed'));
    }
  };

  /**
   * 文件树选择处理
   */
  const handleTreeItemSelect = (record: FileTreeItem) => {
    if (!record) return;
    // For simplified tree, selecting a root folder should navigate to it directly
    if (record.is_dir) {
      // 优化性能：直接设置当前目录而不发起网络请求
      // 这是针对简化版目录树的优化，因为根目录文件夹已经加载过

      // 直接更新状态，避免网络请求导致的延迟
      // 重要：确保设置current以更新pwd，触发params变化，进而刷新右侧文件列表
      store.$patch({
        current: record,
        selected: [record as unknown as FileItem],
        addressItems: [],
      });

      // 通过修改current触发params变化，会自动刷新右侧文件列表
      // 无需手动调用reload()
    } else {
      store.handleTreeItemSelect(record);
    }
  };

  /**
   * 从文件树双击处理
   */
  const handleTreeItemDoubleClick = (record: FileTreeItem) => {
    if (record.is_dir) {
      // 对于目录，导航到该目录
      store.handleOpen(record as unknown as FileItem);
    } else {
      // 获取文件所在目录路径
      const filePath = record.path;
      const lastSlashIndex = filePath.lastIndexOf('/');
      const parentDir =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 检查是否已经在正确的目录
      if (store.pwd !== parentDir) {
        // 如果不在正确的目录，先导航
        store.handleGoto(parentDir);
      }

      // 打开文件
      openFileInEditor(record.path);
    }
  };

  return {
    isGotoTriggered,
    openFileInEditor,
    handleGotoWrapper,
    handleTreeItemSelect,
    handleTreeItemDoubleClick,
  };
};
