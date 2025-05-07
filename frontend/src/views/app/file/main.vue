<template>
  <div class="file-page">
    <a-spin :spinning="loading">
      <address-bar
        :path="store.pwd"
        :items="store.addressItems"
        @search="store.handleAddressSearch"
        @goto="handleGotoWrapper"
      />
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
          @update:show-hidden="updateShowHidden"
          @clear-selected="clearSelected"
          @create="handleCreate"
          @upload="handleUpload"
          @paste="handlePaste"
          @back="store.handleBack"
          @compress="handleBatchCompress"
          @decompress="handleBatchDecompress"
          @delete="handleBatchDelete"
          @terminal="handleTerminal"
          @item-select="handleItemSelect"
          @item-double-click="handleItemDoubleClick"
          @modify-mode="handleModifyMode"
          @modify-owner="handleModifyOwner"
          @operation="handleOperation"
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
  import { computed, inject, onMounted, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/hooks/loading';
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

  // 类型导入
  import { FileItem } from '@/components/file/file-editor-drawer/types';

  // 导入重构的组件
  import SimplifiedFileSidebar from './components/simplified-file-sidebar.vue';
  import FileMainView from './components/file-main-view.vue';

  // 导入组合函数
  import useFileStore from './store/file-store';
  import { useFileOperations } from './hooks/use-file-operations';
  import { useFileNavigation } from './hooks/use-file-navigation';
  import { useFileSelection } from './hooks/use-file-selection';
  import { useFileColumns } from './hooks/use-file-columns';

  const { t } = useI18n();
  const openTerminal = inject<() => void>('openTerminal');
  const { loading, setLoading } = useLoading(false);

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

  // 计算属性
  const showHidden = computed(() => store.showHidden);
  const updateShowHidden = (val: boolean) => {
    store.$patch({
      showHidden: val,
    });
  };

  // 组合函数设置
  const {
    handleModifyMode,
    handleModifyOwner,
    handleCreate,
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
    fileEditorDrawerRef,
    openTerminal,
    selected,
  });

  const { handleGotoWrapper, handleTreeItemSelect, handleTreeItemDoubleClick } =
    useFileNavigation({
      store,
      fileEditorDrawerRef,
    });

  const { handleItemSelect, handleItemDoubleClick, clearSelected } =
    useFileSelection({
      store,
      fileEditorDrawerRef,
      fileMainViewRef,
    });

  const { columns } = useFileColumns(t);

  // 表格参数
  const params = computed(() => {
    return {
      show_hidden: showHidden.value,
      path: store.pwd,
      order_by: 'name',
      order: 'asc',
    } as const;
  });

  // 侦听器
  watch(
    () => store.current,
    (newValue) => {
      if (!newValue) return;
      if (newValue.is_dir) {
        store.handleOpen(newValue as unknown as FileItem);
      }
    }
  );

  watch(params, () => {
    fileMainViewRef.value?.load(params.value);
  });

  // 方法
  const reload = () => {
    fileMainViewRef.value?.reload();
  };

  const handleOk = () => {
    clearSelected();
    reload();
  };

  onMounted(() => {
    store.initTree();
  });
</script>

<style scoped>
  .file-layout {
    position: relative;
    min-height: calc(100vh - 240px);
    margin-top: 20px;
    padding-left: 240px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }
</style>
