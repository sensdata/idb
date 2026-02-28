<template>
  <div class="file-page">
    <a-spin :spinning="loading">
      <div class="address-bar-wrapper">
        <address-bar
          :path="store.pwd"
          :items="store.addressItems"
          @search="store.handleAddressSearch"
          @goto="handleGotoWrapper"
        />
      </div>
      <div class="file-layout">
        <simplified-file-sidebar
          :items="tree"
          :show-hidden="showHidden"
          :current="current"
          @item-select="handleTreeItemSelect"
          @item-double-click="handleTreeItemDoubleClick"
        />
        <file-main-view
          ref="fileMainViewRef"
          :loading="loading"
          :params="params"
          :columns="columns"
          :selected="selected"
          :show-hidden="showHidden"
          :paste-visible="pasteVisible"
          :decompress-visible="decompressVisible"
          :clipboard-count="store.clipboardItems.length"
          @update:show-hidden="updateShowHidden"
          @clear-selected="clearSelected"
          @create="handleCreate"
          @upload="handleUpload"
          @copy="handleCopy"
          @cut="handleCut"
          @paste="handlePasteWrapper"
          @back="handleBack"
          @compress="handleBatchCompress"
          @decompress="handleBatchDecompress"
          @delete="handleBatchDelete"
          @terminal="handleTerminal"
          @item-select="handleItemSelect"
          @item-double-click="handleItemDoubleClick"
          @modify-mode="handleModifyMode"
          @modify-owner="handleModifyOwner"
          @operation="handleOperation"
          @selected-change="handleTableSelectionChange"
          @reload="reload"
        />
      </div>
    </a-spin>

    <!-- Drawers and Modals -->
    <mode-drawer ref="modeDrawerRef" @ok="handleOk" />
    <owner-drawer ref="ownerDrawerRef" @ok="handleOk" />
    <create-file-drawer ref="createFileDrawerRef" @ok="handleOk" />
    <create-folder-drawer ref="createFolderDrawerRef" @ok="handleOk" />
    <rename-drawer ref="renameDrawerRef" @ok="handleOk" />
    <property-drawer ref="propertyDrawerRef" />
    <delete-file-modal ref="deleteFileModalRef" @ok="handleOk" />
    <upload-files-drawer ref="uploadFilesDrawerRef" @ok="handleOk" />
    <compress-drawer ref="compressDrawerRef" @ok="handleOk" />
    <decompress-drawer ref="decompressDrawerRef" @ok="handleOk" />
    <file-editor-drawer ref="fileEditorDrawerRef" @ok="handleOk" />
  </div>
</template>

<script lang="ts" setup>
  import { storeToRefs } from 'pinia';
  import { computed, inject, onMounted, ref, nextTick, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute, useRouter } from 'vue-router';
  import useLoading from '@/composables/loading';
  import { useLogger } from '@/composables/use-logger';
  import {
    parseFilePathFromRoute,
    parsePaginationFromRoute,
    createFileRouteWithPagination,
  } from '@/utils/file-route';
  import useCurrentHost from '@/composables/current-host';
  import AddressBar from '@/components/file/address-bar/index.vue';
  import CreateFileDrawer from '@/components/file/create-file-drawer/index.vue';
  import CreateFolderDrawer from '@/components/file/create-folder-drawer/index.vue';
  import DeleteFileModal from '@/components/file/delete-file-modal/index.vue';
  import ModeDrawer from '@/components/file/mode-drawer/index.vue';
  import OwnerDrawer from '@/components/file/owner-drawer/index.vue';
  import RenameDrawer from '@/components/file/rename-drawer/index.vue';
  import PropertyDrawer from '@/components/file/property-drawer/index.vue';
  import UploadFilesDrawer from '@/components/file/upload-files-drawer/index.vue';
  import CompressDrawer from '@/components/file/compress-drawer/index.vue';
  import DecompressDrawer from '@/components/file/decompress-drawer/index.vue';
  import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';

  // 导入重构的组件
  import SimplifiedFileSidebar from './components/simplified-file-sidebar.vue';
  import FileMainView from './components/file-main-view.vue';

  // 导入组合函数
  import useFileStore from './store/file-store';
  import { useFileOperations } from './composables/use-file-operations';
  import { useFileNavigation } from './composables/use-file-navigation';
  import { useFileSelection } from './composables/use-file-selection';
  import { useFileColumns } from './composables/use-file-columns';

  const { t } = useI18n();
  const route = useRoute();
  const router = useRouter();
  const { currentHostId } = useCurrentHost();
  const openTerminal = inject<() => void>('openTerminal');
  const { loading, setLoading } = useLoading(true);

  // 日志工具
  const { logWarn, logError } = useLogger('FileMain');

  // 组件引用
  const fileMainViewRef = ref<InstanceType<typeof FileMainView>>();
  const modeDrawerRef = ref<InstanceType<typeof ModeDrawer>>();
  const ownerDrawerRef = ref<InstanceType<typeof OwnerDrawer>>();
  const createFileDrawerRef = ref<InstanceType<typeof CreateFileDrawer>>();
  const createFolderDrawerRef = ref<InstanceType<typeof CreateFolderDrawer>>();
  const renameDrawerRef = ref<InstanceType<typeof RenameDrawer>>();
  const propertyDrawerRef = ref<InstanceType<typeof PropertyDrawer>>();
  const deleteFileModalRef = ref<InstanceType<typeof DeleteFileModal>>();
  const uploadFilesDrawerRef = ref<InstanceType<typeof UploadFilesDrawer>>();
  const compressDrawerRef = ref<InstanceType<typeof CompressDrawer>>();
  const decompressDrawerRef = ref<InstanceType<typeof DecompressDrawer>>();
  const fileEditorDrawerRef = ref<InstanceType<typeof FileEditorDrawer>>();

  // 存储设置
  const store = useFileStore();
  const { current, tree, pasteVisible, decompressVisible, selected } =
    storeToRefs(store);

  // 直接从 store 解构
  const { showHidden } = storeToRefs(store);
  const updateShowHidden = (val: boolean) => {
    store.showHidden = val;
  };

  // 组合函数设置
  const {
    handleModifyMode,
    handleModifyOwner,
    handleCreate,
    handleCopy,
    handleCut,
    handleBatchCompress,
    handleBatchDecompress,
    handleBatchDelete,
    handleOperation,
    handlePaste,
    handleUpload,
    handleTerminal,
  } = useFileOperations({
    store,
    setLoading,
    modeDrawerRef,
    ownerDrawerRef,
    createFileDrawerRef,
    createFolderDrawerRef,
    compressDrawerRef,
    decompressDrawerRef,
    deleteFileModalRef,
    uploadFilesDrawerRef,
    renameDrawerRef,
    propertyDrawerRef,
    fileEditorDrawerRef,
    openTerminal,
    selected,
  });

  // 表格参数（包含从URL解析的分页参数）
  const params = computed(() => {
    const paginationParams = parsePaginationFromRoute(route.query);
    // 直接从路由查询参数解析路径，避免依赖store.pwd的时序问题
    const currentPath = parseFilePathFromRoute(route.query);

    const result = {
      show_hidden: showHidden.value,
      path: currentPath, // 使用从路由解析的路径
      order_by: 'name',
      order: 'asc',
      page: paginationParams.page,
      page_size: paginationParams.pageSize,
    } as const;

    return result;
  });

  const { handleGotoWrapper, handleTreeItemSelect, handleTreeItemDoubleClick } =
    useFileNavigation({
      store,
      fileEditorDrawerRef,
      router,
      currentHostId,
      setLoading,
    });

  const { handleItemSelect, handleItemDoubleClick, clearSelected } =
    useFileSelection({
      store,
      fileEditorDrawerRef,
      fileMainViewRef,
      router,
      currentHostId,
      setLoading,
    });

  const { columns } = useFileColumns(t);

  // 方法
  const reload = () => {
    fileMainViewRef.value?.reload();
  };

  const handleOk = () => {
    clearSelected();
    reload();
  };

  // 获取父级路径
  const getParentPath = (currentPath: string): string => {
    if (currentPath === '/') return '/';

    const normalizedPath =
      currentPath.endsWith('/') && currentPath !== '/'
        ? currentPath.slice(0, -1)
        : currentPath;

    const lastSlashIndex = normalizedPath.lastIndexOf('/');
    return lastSlashIndex <= 0 ? '/' : normalizedPath.slice(0, lastSlashIndex);
  };

  // 处理返回按钮
  const handleBack = () => {
    const currentPath = store.pwd;

    // 检查是否可以返回
    if (!store.current || currentPath === '/') {
      return;
    }

    setLoading(true);

    const parentPath = getParentPath(currentPath);
    const routeConfig = createFileRouteWithPagination(
      parentPath,
      undefined, // 重置分页参数
      currentHostId.value ? { id: currentHostId.value } : {}
    );

    router.push(routeConfig);
  };

  // 处理表格选择变化
  const handleTableSelectionChange = (selectedItems: any[]) => {
    // 更新store中的选中状态，这样工具栏就会正确显示
    store.handleSelected(selectedItems);
  };

  // 包装粘贴处理，成功后刷新目录
  const handlePasteWrapper = async () => {
    const success = await handlePaste();
    if (success) {
      reload();
    }
  };

  onMounted(async () => {
    try {
      // 初始化文件树
      store.initTree();

      // 等待路径导航完成
      const targetPath = parseFilePathFromRoute(route.query);

      if (targetPath !== store.pwd) {
        // 如果当前路径与目标路径不同，等待导航完成
        await store.navigateToPath(targetPath);
      }

      // 路径导航完成后，重置loading状态，然后手动触发表格加载
      await nextTick();

      // 重置loading状态，避免表格加载被跳过
      setLoading(false);
      await nextTick();

      // 简化数据加载逻辑
      if (fileMainViewRef.value?.load) {
        await fileMainViewRef.value.load(params.value);
      } else {
        logWarn('File view ref not available during mount');
        // 等待下一个tick后再试一次
        await nextTick();
        if (fileMainViewRef.value?.load) {
          await fileMainViewRef.value.load(params.value);
        } else {
          logError('File view ref still not available after nextTick');
        }
      }
    } catch (error) {
      logError('Failed to initialize file view:', error);
      setLoading(false);
    }
  });

  // 监听路径查询参数变化
  watch(
    () => parseFilePathFromRoute(route.query),
    async (newPath, oldPath) => {
      // 只在路径真正变化时处理
      if (newPath !== oldPath && newPath !== store.pwd) {
        setLoading(true);

        try {
          await store.navigateToPath(newPath);
          await nextTick();

          // 重置loading状态，避免表格加载被跳过
          setLoading(false);
          await nextTick();

          if (fileMainViewRef.value?.load) {
            await fileMainViewRef.value.load(params.value);
          } else {
            logWarn('File view ref not available after path change');
          }
        } catch (error) {
          logError('Failed to load after path change:', error);
          setLoading(false);
        }
      }
    }
  );
</script>

<style scoped>
  .file-page {
    display: flex;
    flex-direction: column;
    width: 100%;
  }

  .address-bar-wrapper {
    width: 100%;
    margin-bottom: 20px;
  }

  .file-layout {
    display: flex;
    align-items: stretch;
    min-height: calc(100vh - 240px);
    overflow: hidden;
    background: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
  }
</style>
