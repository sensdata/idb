<template>
  <div class="file-page">
    <a-spin :spinning="loading">
      <address-bar
        :path="store.pwd"
        :items="store.addressItems"
        @search="store.handleAddressSearch"
        @goto="store.handleGoto"
      />
      <div class="file-layout">
        <div class="file-sidebar">
          <file-tree
            :items="tree"
            :show-hidden="showHidden"
            :selected="current"
            :selected-change="store.handleTreeItemSelect"
            :open-change="store.handleTreeItemOpenChange"
          />
        </div>
        <div class="file-main">
          <div class="action-wrap mb-4">
            <a-button-group class="idb-button-group">
              <a-button @click="store.handleBack">
                <icon-left />
                <span class="ml-2">{{ t('app.file.list.action.back') }}</span>
              </a-button>
              <template v-if="!selected.length">
                <a-dropdown position="bl" @select="handleCreate">
                  <a-button>
                    <icon-plus />
                    <span class="mx-2">{{
                      $t('app.file.list.action.create')
                    }}</span>
                    <icon-caret-down />
                  </a-button>

                  <template #content>
                    <a-doption value="createFolder">
                      <template #icon>
                        <icon-plus />
                      </template>
                      <template #default>
                        <span class="ml-2">{{
                          $t('app.file.list.action.createFolder')
                        }}</span>
                      </template>
                    </a-doption>
                    <a-doption value="createFile">
                      <template #icon>
                        <icon-plus />
                      </template>
                      <template #default>
                        <span class="ml-2">{{
                          $t('app.file.list.action.createFile')
                        }}</span>
                      </template>
                    </a-doption>
                  </template>
                </a-dropdown>
                <a-button @click="handleUpload">
                  <icon-upload />
                  <span class="ml-2">{{
                    $t('app.file.list.action.upload')
                  }}</span>
                </a-button>
              </template>
              <template v-else>
                <a-button @click="store.handleCopy">
                  <icon-copy />
                  <span class="ml-2">{{
                    $t('app.file.list.action.copy')
                  }}</span>
                </a-button>
                <a-button @click="store.handleCut">
                  <icon-scissor />
                  <span class="ml-2">{{ $t('app.file.list.action.cut') }}</span>
                </a-button>
                <a-button
                  v-if="decompressVisible"
                  @click="handleBatchDecompress"
                >
                  <decompression-icon />
                  <span class="ml-2">{{
                    $t('app.file.list.action.decompress')
                  }}</span>
                </a-button>
                <a-button v-else @click="handleBatchCompress">
                  <compression-icon />
                  <span class="ml-2">{{
                    $t('app.file.list.action.compress')
                  }}</span>
                </a-button>
                <a-button @click="handleBatchDelete">
                  <icon-delete />
                  <span class="ml-2">{{
                    $t('app.file.list.action.delete')
                  }}</span>
                </a-button>
              </template>
              <a-button @click="handleTerminal">
                <icon-code-square />
                <span class="ml-2">{{
                  $t('app.file.list.action.terminal')
                }}</span>
              </a-button>
            </a-button-group>
            <a-button-group v-if="pasteVisible" class="idb-button-group ml-4">
              <a-button @click="handlePaste">
                <icon-paste />
                <span class="ml-2">
                  {{
                    $t('app.file.list.action.paste', { count: selected.length })
                  }}
                </span>
              </a-button>
              <a-button @click="clearSelected">
                <template #icon>
                  <icon-close />
                </template>
              </a-button>
            </a-button-group>
          </div>
          <idb-table
            ref="gridRef"
            :params="params"
            :columns="columns"
            :fetch="getFileListApi"
            has-batch
            row-key="path"
            @selected-change="store.handleSelected"
          >
            <template #leftActions>
              <a-checkbox v-model="showHidden">{{
                $t('app.file.list.filter.showHidden')
              }}</a-checkbox>
            </template>
            <template #name="{ record }: { record: FileItem }">
              <div
                class="name-cell flex items-center"
                @click="store.handleOpen(record)"
              >
                <folder-icon v-if="record.is_dir" />
                <file-icon v-else />
                <span
                  class="color-primary cursor-pointer min-w-0 flex-1 truncate"
                  >{{ record.name }}</span
                >
              </div>
            </template>
            <template #mode="{ record }: { record: FileItem }">
              <div
                class="color-primary cursor-pointer"
                @click="handleModifyMode(record)"
                >{{ record.mode }}</div
              >
            </template>
            <template #user="{ record }: { record: FileItem }">
              <div
                class="color-primary cursor-pointer"
                @click="handleModifyOwner(record)"
                >{{ record.user }}</div
              >
            </template>
            <template #group="{ record }: { record: FileItem }">
              <div
                class="color-primary cursor-pointer"
                @click="handleModifyOwner(record)"
                >{{ record.group }}</div
              >
            </template>
            <template #operation="{ record }: { record: FileItem }">
              <a-dropdown
                :popup-max-height="false"
                @select="handleOperation($event, record)"
              >
                <a-button type="text">
                  <icon-settings />
                  <icon-caret-down class="ml-4" />
                </a-button>
                <template #content>
                  <a-doption value="open">
                    {{ $t('app.file.list.operation.open') }}
                  </a-doption>
                  <a-doption value="mode">
                    {{ $t('app.file.list.operation.mode') }}
                  </a-doption>
                  <a-doption value="rename">
                    {{ $t('app.file.list.operation.rename') }}
                  </a-doption>
                  <a-doption value="copyPath">
                    {{ $t('app.file.list.operation.copyPath') }}
                  </a-doption>
                  <a-doption value="download">
                    {{ $t('app.file.list.operation.download') }}
                  </a-doption>
                  <a-doption value="property">
                    {{ $t('app.file.list.operation.property') }}
                  </a-doption>
                  <a-doption value="delete">
                    {{ $t('app.file.list.operation.delete') }}
                  </a-doption>
                </template>
              </a-dropdown>
            </template>
          </idb-table>
        </div>
      </div>
    </a-spin>
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
  </div>
</template>

<script lang="ts" setup>
  import { storeToRefs } from 'pinia';
  import {
    computed,
    GlobalComponents,
    inject,
    onMounted,
    ref,
    unref,
    watch,
  } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { resolveApiUrl } from '@/helper/api-helper';
  import { getFileListApi, moveFileApi } from '@/api/file';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { formatFileSize, formatTime } from '@/utils/format';
  import { useClipboard } from '@/hooks/use-clipboard';
  import useLoading from '@/hooks/loading';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import CompressionIcon from '@/assets/icons/compression.svg';
  import DecompressionIcon from '@/assets/icons/decompression.svg';
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
  import FileTree from './components/file-tree/index.vue';
  import useFileStore from './store/file-store';
  import { FileItem } from './types/file-item';

  const { t } = useI18n();
  const openTerminal = inject<() => void>('openTerminal');
  const { copyText } = useClipboard();
  const { loading, setLoading } = useLoading(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
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
  const store = useFileStore();
  const { current, tree, pasteVisible, decompressVisible, selected } =
    storeToRefs(store);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.file.list.column.name'),
      width: 200,
      ellipsis: true,
      slotName: 'name',
    },
    {
      dataIndex: 'size',
      title: t('app.file.list.column.size'),
      width: 100,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatFileSize(record.size);
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.file.list.column.mod_time'),
      width: 180,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      dataIndex: 'mode',
      title: t('app.file.list.column.mode'),
      width: 100,
      slotName: 'mode',
    },
    {
      dataIndex: 'user',
      title: t('app.file.list.column.user'),
      width: 100,
      slotName: 'user',
    },
    {
      dataIndex: 'group',
      title: t('app.file.list.column.group'),
      width: 100,
      slotName: 'group',
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const showHidden = computed({
    get: () => store.showHidden,
    set: (val) =>
      store.$patch({
        showHidden: val,
      }),
  });
  const params = computed(() => {
    return {
      show_hidden: showHidden.value,
      path: store.pwd,
    };
  });
  watch(params, () => {
    gridRef.value?.load(params.value);
  });

  const clearSelected = () => {
    store.clearSelected();
    gridRef.value?.clearSelected();
  };

  const handleCreate = (key: any) => {
    switch (key) {
      case 'createFolder':
        createFolderDrawerRef.value?.setData({
          pwd: store.pwd,
        });
        createFolderDrawerRef.value?.show();
        break;
      case 'createFile':
        createFileDrawerRef.value?.setData({
          pwd: store.pwd,
        });
        createFileDrawerRef.value?.show();
        break;
      default:
        break;
    }
  };

  const handleModifyMode = (record: FileItem) => {
    modeDrawerRef.value?.setData(record);
    modeDrawerRef.value?.show();
  };

  const handleModifyOwner = (record: FileItem) => {
    ownerDrawerRef.value?.setData(record);
    ownerDrawerRef.value?.show();
  };

  const handlePaste = async () => {
    setLoading(true);
    try {
      await moveFileApi({
        sources: store.selected.map((item) => item.path),
        dest: store.pwd,
        cover: false,
        type: store.cutActive ? 'cut' : 'copy',
      });
      clearSelected();
      gridRef.value?.reload();
    } finally {
      setLoading(false);
    }
  };

  const handleUpload = () => {
    uploadFilesDrawerRef.value?.show();
    uploadFilesDrawerRef.value?.setData({
      directory: store.pwd,
    });
  };
  const handleTerminal = () => {
    // const pwd = store.pwd;
    // todo: set terminal pwd path
    openTerminal?.();
  };

  const handleBatchCompress = () => {
    if (!selected.value?.length) {
      return;
    }
    compressDrawerRef.value?.setFiles(unref(selected));
    compressDrawerRef.value?.show();
  };

  const handleBatchDecompress = () => {
    if (!selected.value?.length) {
      return;
    }
    decompressDrawerRef.value?.setFiles(unref(selected));
    decompressDrawerRef.value?.show();
  };

  const handleBatchDelete = () => {
    if (!selected.value?.length) {
      return;
    }
    deleteFileModalRef.value?.setData(unref(selected));
    deleteFileModalRef.value?.show();
  };

  const handleRename = (record: FileItem) => {
    renameDrawerRef.value?.setData({
      path: record.path,
    });
    renameDrawerRef.value?.show();
  };

  const handleCopyPath = async (record: FileItem) => {
    try {
      await copyText(record.path);
      Message.success(t('app.file.list.message.copyPathSuccess'));
    } catch (err) {
      Message.error(t('app.file.list.message.copyPathFailed'));
    }
  };

  const handleDownload = (record: FileItem) => {
    const a = document.createElement('a');
    a.href = resolveApiUrl('/files/{host}/download', { source: record.path });
    a.download = record.name;
    a.click();
  };

  const handleProperty = (record: FileItem) => {
    propertyDrawerRef.value?.setData(record);
    propertyDrawerRef.value?.show();
  };

  const handleDelete = (record: FileItem) => {
    deleteFileModalRef.value?.setData([record]);
    deleteFileModalRef.value?.show();
  };

  const handleOperation = (key: any, record: FileItem) => {
    switch (key) {
      case 'open':
        store.handleOpen(record);
        break;
      case 'mode':
        handleModifyMode(record);
        break;
      case 'rename':
        handleRename(record);
        break;
      case 'copyPath':
        handleCopyPath(record);
        break;
      case 'download':
        handleDownload(record);
        break;
      case 'property':
        handleProperty(record);
        break;
      case 'delete':
        handleDelete(record);
        break;
      default:
        break;
    }
  };

  const reload = () => {
    gridRef.value?.reload();
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

  .file-sidebar {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 240px;
    height: 100%;
    padding: 4px 8px;
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
  }

  .file-main {
    min-width: 0;
    height: 100%;
    padding: 20px;
  }

  .name-cell svg {
    width: 14px;
    height: 14px;
    margin-right: 8px;
    vertical-align: top;
  }
</style>
