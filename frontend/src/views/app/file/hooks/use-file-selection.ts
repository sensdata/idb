import { Ref, watch, nextTick } from 'vue';
import { debounce } from 'lodash';
import { getFileListApi, getFileDetailApi } from '@/api/file';
import { FileInfoEntity } from '@/entity/FileInfo';
import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';
import { resolveApiUrl } from '@/helper/api-helper';
import { openWindow } from '@/utils';
import { FileItem } from '../types/file-item';
import useFileStore from '../store/file-store';
import FileMainView from '../components/file-main-view.vue';

interface FileSelectionParams {
  store: ReturnType<typeof useFileStore>;
  fileEditorDrawerRef: Ref<InstanceType<typeof FileEditorDrawer> | undefined>;
  fileMainViewRef: Ref<InstanceType<typeof FileMainView> | undefined>;
}

export const useFileSelection = (params: FileSelectionParams) => {
  const { store, fileEditorDrawerRef, fileMainViewRef } = params;

  // 引用导航模块中的函数
  const openFileInEditor = async (fileOrPath: FileItem | string) => {
    try {
      const filePath =
        typeof fileOrPath === 'string' ? fileOrPath : fileOrPath.path;

      // 提取目录路径
      const lastSlashIndex = filePath.lastIndexOf('/');
      const dirPath =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 如果不在当前目录，先导航
      if (store.pwd !== dirPath) {
        store.handleGoto(dirPath);
      }

      // 获取文件详情 - 直接使用API而不是store方法
      const fileDetail = await getFileDetailApi({
        path: filePath,
        expand: false,
      });

      if (!fileDetail) return;

      // 根据文件大小决定打开方式
      if (fileDetail.size > 1048576) {
        // 大文件下载 - 使用openWindow工具函数而不是DOM操作
        const downloadUrl = resolveApiUrl('/files/{host}/download', {
          source: fileDetail.path,
        });
        openWindow(downloadUrl, { download: fileDetail.name });
      } else {
        // 小文件打开编辑器
        fileEditorDrawerRef.value?.setFile(fileDetail);
        fileEditorDrawerRef.value?.show();
      }
    } catch (error) {
      console.error('File open error:', error);
    }
  };

  /**
   * 单击处理：目录只选择，文件则打开
   */
  const handleSingleClickAction = (record: FileItem) => {
    if (record.is_dir) {
      // 目录进行选择并打开
      store.handleSelected([record]);
      store.handleOpen(record);
    } else {
      // 文件则打开
      const filePath = record.path;
      const lastSlashIndex = filePath.lastIndexOf('/');
      const parentDir =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 如果不在当前目录，则导航
      if (store.pwd !== parentDir) {
        store.handleGoto(parentDir);
      }

      // 打开文件
      openFileInEditor(record);
    }
  };

  // 使用debounce函数处理单击事件，防止与双击事件冲突
  const handleItemSelect = debounce((record: FileItem) => {
    handleSingleClickAction(record);
  }, 250);

  /**
   * 双击处理：导航并打开
   */
  const handleItemDoubleClick = (record: FileItem) => {
    // 取消单击事件的执行
    handleItemSelect.cancel();

    if (record.is_dir) {
      // 打开目录
      store.handleOpen(record);
    } else {
      // 文件则打开
      const filePath = record.path;
      const lastSlashIndex = filePath.lastIndexOf('/');
      const parentDir =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 如果不在当前目录，则导航
      if (store.pwd !== parentDir) {
        store.handleGoto(parentDir);
      }

      // 打开文件
      openFileInEditor(record);
    }
  };

  // 文件定位和高亮显示
  const findFileAndHighlight = async (fileName: string, parentDir: string) => {
    try {
      // 使用 Vue 的 nextTick 等待 DOM 更新完成
      await nextTick();

      // 获取页面大小
      const pageSize = 20;

      // 获取文件列表计算页码
      const fileListResponse = await getFileListApi({
        path: parentDir,
        show_hidden: store.showHidden,
        order_by: 'name',
        order: 'asc',
        page: 1,
        page_size: 1000, // 足够多的文件
      });

      if (fileListResponse && fileListResponse.items) {
        // 查找文件索引
        const fileIndex = fileListResponse.items.findIndex(
          (item: FileInfoEntity) => item.name === fileName
        );

        if (fileIndex !== -1) {
          // 计算页码
          const page = Math.floor(fileIndex / pageSize) + 1;

          // 导航到页面
          if (fileMainViewRef.value && page) {
            fileMainViewRef.value.load({
              path: parentDir,
              show_hidden: store.showHidden,
              order_by: 'name',
              order: 'asc',
              page,
            });
          }
        }
      }
    } catch (error) {
      console.error('处理文件导航失败:', error);
    }
  };

  // 清除选中项
  const clearSelected = () => {
    store.clearSelected();
    fileMainViewRef.value?.clearSelected();
  };

  // 初始化监听器
  const initSelectionWatchers = () => {
    // 监听选择变化，单击只定位不打开
    watch(
      () => store.selected,
      (newSelected) => {
        // 只处理单文件选择
        if (newSelected.length === 1 && !newSelected[0].is_dir) {
          const selectedFile = newSelected[0];

          // 对于文件，检查我们是否已经在正确的目录中
          const filePath = selectedFile.path;
          const fileName = selectedFile.name;
          const lastSlashIndex = filePath.lastIndexOf('/');
          const parentDir =
            lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

          let fileInView = false;

          // 检查我们是否已经在父目录中
          if (store.pwd === parentDir) {
            // 检查文件是否已经在当前视图中
            const currentFiles = fileMainViewRef.value?.getData?.() || [];
            if (
              currentFiles.length > 0 &&
              currentFiles.some((file: any) => file.name === fileName)
            ) {
              // 文件已经存在于当前视图中，只需选择它而不需要导航
              fileInView = true;
            }
          }

          if (!fileInView) {
            // 执行文件定位
            findFileAndHighlight(fileName, parentDir);
          }
        }
      }
    );
  };

  // 初始化
  initSelectionWatchers();

  return {
    handleItemSelect,
    handleItemDoubleClick,
    clearSelected,
    findFileAndHighlight,
  };
};
