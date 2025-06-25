import { Ref, watch, nextTick } from 'vue';
import { debounce } from 'lodash';
import { Router } from 'vue-router';
import { getFileListApi, getFileDetailApi, getFileTailApi } from '@/api/file';
import { FileInfoEntity } from '@/entity/FileInfo';
import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';
import {
  ContentViewMode,
  FileItem,
} from '@/components/file/file-editor-drawer/types';
import { useLogger } from '@/composables/use-logger';
import { createFileRouteWithPagination } from '@/utils/file-route';
import useFileStore from '../store/file-store';
import FileMainView from '../components/file-main-view.vue';

interface FileSelectionParams {
  store: ReturnType<typeof useFileStore>;
  fileEditorDrawerRef: Ref<InstanceType<typeof FileEditorDrawer> | undefined>;
  fileMainViewRef: Ref<InstanceType<typeof FileMainView> | undefined>;
  router: Router;
  currentHostId: Ref<number | undefined>;
  setLoading?: (loading: boolean) => void;
}

export const useFileSelection = (params: FileSelectionParams) => {
  const {
    store,
    fileEditorDrawerRef,
    fileMainViewRef,
    router,
    currentHostId,
    setLoading,
  } = params;
  const { logError, logDebug } = useLogger('FileSelection');

  // 引用导航模块中的函数
  const openFileInEditor = async (fileOrPath: FileItem | string) => {
    try {
      const filePath =
        typeof fileOrPath === 'string' ? fileOrPath : fileOrPath.path;

      // 先获取基本信息以确定文件名
      const fileName =
        typeof fileOrPath === 'string'
          ? filePath.substring(filePath.lastIndexOf('/') + 1)
          : fileOrPath.name;

      // 立即打开编辑器，先显示加载状态
      const initialFileInfo: Partial<FileItem> = {
        path: filePath,
        name: fileName,
        content: '',
        loading: true,
        content_view_mode: 'loading',
        // 添加必要的基本信息
        size: 0,
        is_dir: false,
        mode: '',
        user: '',
        group: '',
        mod_time: '',
        extension: '',
        favorite_id: 0,
        gid: '',
        is_hidden: false,
        is_symlink: false,
        item_total: 0,
        link_path: '',
        mime_type: '',
        type: '',
        uid: '',
        update_time: '',
      };

      fileEditorDrawerRef.value?.setFile(initialFileInfo as FileItem);
      fileEditorDrawerRef.value?.show();

      // 提取目录路径
      const lastSlashIndex = filePath.lastIndexOf('/');
      const dirPath =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 如果不在当前目录，先导航
      if (store.pwd !== dirPath) {
        store.handleGoto(dirPath);
      }

      // 延迟500ms，确保能看到loading效果
      await new Promise((resolve) => {
        setTimeout(resolve, 500);
      });

      // 获取文件详情 - 直接使用API而不是store方法
      const fileDetail = await getFileDetailApi({
        path: filePath,
        expand: false,
      });

      if (!fileDetail) return;

      // 默认显示行数
      const defaultLineCount = 30;

      // 根据文件大小决定打开方式
      if (fileDetail.size > 100000) {
        // 文件大于100K，使用tail API获取最后1000行
        const tailData = await getFileTailApi({
          path: filePath,
          numbers: defaultLineCount,
        });

        // 创建一个包含尾部内容的文件详情对象
        const fileWithPartialContent = {
          ...fileDetail,
          content: tailData.content,
          is_tail: true, // 保留向后兼容
          content_view_mode: 'tail' as ContentViewMode,
          line_count: defaultLineCount,
          loading: false,
        };

        // 更新编辑器内容
        fileEditorDrawerRef.value?.setFile(fileWithPartialContent);
      } else {
        // 小文件完整打开编辑器，但仍然允许使用实时追踪模式
        fileEditorDrawerRef.value?.setFile({
          ...fileDetail,
          content_view_mode: 'full',
          loading: false,
        });
      }
    } catch (error) {
      logError('File open error:', error);
      // 发生错误时关闭编辑器
      fileEditorDrawerRef.value?.hide();
    }
  };

  /**
   * 单击处理：目录只选择，文件则打开
   */
  const handleSingleClickAction = (record: FileItem) => {
    if (record.is_dir) {
      // 立即设置loading状态，避免显示空白页面
      if (setLoading) {
        logDebug(
          '🔄 useFileSelection handleSingleClickAction: setting loading to true immediately'
        );
        setLoading(true);
      }

      // 目录进行选择并打开，使用路由导航
      store.handleSelected([record]);

      // 创建新的路由配置，不传递分页参数以重置page为1
      const routeConfig = createFileRouteWithPagination(
        record.path,
        undefined, // 不传递分页参数，让page重置为默认值
        currentHostId.value ? { id: currentHostId.value } : {}
      );

      logDebug('🚀 useFileSelection handleSingleClickAction navigation:', {
        targetPath: record.path,
        routeConfig,
        currentPath: store.pwd,
      });

      // 使用push导航，更新URL并重置分页
      router.push(routeConfig);

      // 同时也调用store方法更新内部状态
      store.handleOpen(record);
    } else {
      // 文件进行选择
      store.handleSelected([record]);

      // 直接打开文件，不再依赖导航状态
      openFileInEditor(record);
    }
  };

  // 降低debounce延迟，确保单击响应更快速，同时仍然能够与双击区分开
  const handleItemSelect = debounce((record: FileItem) => {
    handleSingleClickAction(record);
  }, 150);

  /**
   * 双击处理：导航并打开
   */
  const handleItemDoubleClick = (record: FileItem) => {
    // 取消单击事件的执行
    handleItemSelect.cancel();

    if (record.is_dir) {
      // 立即设置loading状态，避免显示空白页面
      if (setLoading) {
        logDebug(
          '🔄 useFileSelection handleItemDoubleClick: setting loading to true immediately'
        );
        setLoading(true);
      }

      // 创建新的路由配置，不传递分页参数以重置page为1
      const routeConfig = createFileRouteWithPagination(
        record.path,
        undefined, // 不传递分页参数，让page重置为默认值
        currentHostId.value ? { id: currentHostId.value } : {}
      );

      logDebug('🚀 useFileSelection handleItemDoubleClick navigation:', {
        targetPath: record.path,
        routeConfig,
        currentPath: store.pwd,
      });

      // 使用push导航，更新URL并重置分页
      router.push(routeConfig);

      // 同时也调用store方法更新内部状态
      store.handleOpen(record);
    } else {
      // 直接打开文件，不需要额外的导航检查
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
      logError('处理文件导航失败:', error);
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
