<template>
  <div class="file-main">
    <div class="toolbar">
      <a-button-group class="idb-button-group">
        <a-button @click="$emit('back')">
          <icon-left />
        </a-button>
        <a-dropdown
          position="bl"
          @select="(key) => handleCreate(key as string)"
        >
          <a-button>
            <icon-plus />
            <span class="mx-2">{{ $t('app.file.list.action.create') }}</span>
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
        <a-button @click="$emit('upload')">
          <icon-upload />
          <span class="ml-2">{{ $t('app.file.list.action.upload') }}</span>
        </a-button>

        <a-button @click="$emit('terminal')">
          <icon-code-square />
          <span class="ml-2">{{ $t('app.file.list.action.terminal') }}</span>
        </a-button>
      </a-button-group>

      <a-button-group
        v-if="selectedItems.length > 0"
        class="idb-button-group ml-4"
      >
        <a-button @click="$emit('copy')">
          <icon-copy />
          <span class="ml-2">{{ $t('app.file.list.action.copy') }}</span>
        </a-button>
        <a-button @click="$emit('cut')">
          <icon-scissor />
          <span class="ml-2">{{ $t('app.file.list.action.cut') }}</span>
        </a-button>
        <a-button v-if="decompressVisible" @click="$emit('decompress')">
          <decompression-icon />
          <span class="ml-2">{{ $t('app.file.list.action.decompress') }}</span>
        </a-button>
        <a-button v-else @click="$emit('compress')">
          <compression-icon />
          <span class="ml-2">{{ $t('app.file.list.action.compress') }}</span>
        </a-button>
        <a-button @click="$emit('delete')">
          <icon-delete />
          <span class="ml-2">{{ $t('app.file.list.action.delete') }}</span>
        </a-button>
      </a-button-group>

      <a-button-group v-if="pasteVisible" class="idb-button-group ml-4">
        <a-button @click="$emit('paste')">
          <icon-paste />
          <span class="ml-2">
            {{
              $t('app.file.list.action.paste', { count: selectedItems.length })
            }}
          </span>
        </a-button>
        <a-button @click="$emit('clearSelected')">
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
      :loading="loading"
      has-batch
      row-key="path"
      url-sync
      :auto-load="false"
      @selected-change="handleTableSelectionChange"
    >
      <template #leftActions>
        <a-checkbox
          :modelValue="props.showHidden"
          @update:model-value="handleShowHiddenChange"
        >
          {{ $t('app.file.list.filter.showHidden') }}
        </a-checkbox>
      </template>
      <template #name="{ record }">
        <div
          class="name-cell flex items-center"
          @click="handleItemSelect(record)"
          @dblclick="handleItemDoubleClick(record)"
        >
          <folder-icon v-if="record.is_dir && !record.is_symlink" />
          <file-icon v-if="!record.is_dir && !record.is_symlink" />
          <icon-link v-if="record.is_symlink" style="color: #1890ff" />
          <span class="color-primary cursor-pointer min-w-0 flex-1 truncate">
            {{ record.name }}
            <template v-if="record.is_symlink">
              <span class="text-gray-500 ml-1">→</span>
              <span class="text-gray-500 ml-1 italic">{{
                record.link_path
              }}</span>
            </template>
          </span>
        </div>
      </template>
      <template #mode="{ record }">
        <div
          class="color-primary cursor-pointer"
          @click="$emit('modifyMode', record)"
          >{{ record.mode }}</div
        >
      </template>
      <template #user="{ record }">
        <div
          class="color-primary cursor-pointer"
          @click="$emit('modifyOwner', record)"
          >{{ record.user }}</div
        >
      </template>
      <template #group="{ record }">
        <div
          class="color-primary cursor-pointer"
          @click="$emit('modifyOwner', record)"
          >{{ record.group }}</div
        >
      </template>
      <template #operation="{ record }">
        <a-dropdown
          :popup-max-height="false"
          @select="(key) => handleOperation(key as string, record)"
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
</template>

<script lang="ts" setup>
  import { ref, GlobalComponents, watch, computed } from 'vue';
  import { debounce } from 'lodash';
  import { getFileListApi } from '@/api/file';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import CompressionIcon from '@/assets/icons/compression.svg';
  import DecompressionIcon from '@/assets/icons/decompression.svg';
  import { FileItem } from '@/components/file/file-editor-drawer/types';
  import { useLogger } from '@/composables/use-logger';

  interface TableColumn {
    dataIndex: string;
    title: string;
    width?: number;
    ellipsis?: boolean;
    slotName?: string;
    render?: (data: any) => any;
    align?: 'left' | 'center' | 'right';
  }

  const props = defineProps<{
    params: Record<string, any>;
    columns: TableColumn[];
    loading?: boolean;
    showHidden?: boolean;
    selected: FileItem[];
    pasteVisible?: boolean;
    decompressVisible?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'back'): void;
    (e: 'upload'): void;
    (e: 'copy'): void;
    (e: 'cut'): void;
    (e: 'paste'): void;
    (e: 'clearSelected'): void;
    (e: 'terminal'): void;
    (e: 'create', type: string): void;
    (e: 'compress'): void;
    (e: 'decompress'): void;
    (e: 'delete'): void;
    (e: 'reload'): void;
    (e: 'selectedChange', selected: FileItem[]): void;
    (e: 'update:showHidden', value: boolean): void;
    (e: 'itemSelect', record: FileItem): void;
    (e: 'itemDoubleClick', record: FileItem): void;
    (e: 'modifyMode', record: FileItem): void;
    (e: 'modifyOwner', record: FileItem): void;
    (e: 'operation', key: string, record: FileItem): void;
  }>();

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  // 初始化日志
  const { logInfo, logWarn } = useLogger('FileMainView');

  // 直接使用props中的状态
  const selectedItems = computed(() => props.selected || []);

  // 事件处理直接向上传递，不维护内部状态
  const handleTableSelectionChange = (selected: FileItem[]) => {
    emit('selectedChange', selected);
  };

  // Handle item selection with debounce to avoid conflicts with double-click
  const DEBOUNCE_DELAY = 150;
  const handleItemSelect = debounce((record: FileItem) => {
    emit('itemSelect', record);
  }, DEBOUNCE_DELAY);

  const handleItemDoubleClick = (record: FileItem) => {
    // Cancel single click action to prevent conflicts
    handleItemSelect.cancel();
    emit('itemDoubleClick', record);
  };

  // Handle dropdown menu operations
  const handleOperation = (key: string, record: FileItem) => {
    emit('operation', key, record);
  };

  // Handle create folder/file dropdown selection
  const handleCreate = (type: string) => {
    emit('create', type);
  };

  // Public methods exposed for parent component
  const load = async (customParams?: any) => {
    logInfo('load method called:', {
      customParams,
      tableRef: !!gridRef.value,
      hasLoadMethod: !!(gridRef.value && gridRef.value.load),
      receivedLoadingProp: props.loading,
      urlSync: true,
      timestamp: new Date().toISOString(),
    });

    if (gridRef.value?.load) {
      logInfo('calling gridRef.load with params:', customParams);
      await gridRef.value.load(customParams);
      logInfo('gridRef.load completed');
    } else {
      logWarn('gridRef or load method not available:', {
        gridRefExists: !!gridRef.value,
        loadMethodExists: !!(gridRef.value && gridRef.value.load),
      });
    }
  };

  const reload = () => {
    gridRef.value?.reload();
  };

  // 清除选择也直接向上传递
  const clearSelected = () => {
    emit('selectedChange', []);
    gridRef.value?.clearSelected();
  };

  const getData = () => {
    return gridRef.value?.getData();
  };

  defineExpose({
    load,
    reload,
    getData,
    clearSelected,
  });

  // 监听loading状态变化，用于调试
  watch(
    () => props.loading,
    (newLoading, oldLoading) => {
      logInfo('loading prop changed:', {
        from: oldLoading,
        to: newLoading,
        timestamp: new Date().toISOString(),
      });
    },
    { immediate: true }
  );

  // 1. 提取事件处理器
  const handleShowHiddenChange = (
    value: boolean | (string | number | boolean)[]
  ) => {
    emit('update:showHidden', Boolean(value));
  };
</script>

<style scoped>
  .file-main {
    flex: 1;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 20px;
    overflow: auto;
  }

  .name-cell svg {
    width: 14px;
    height: 14px;
    margin-right: 8px;
    vertical-align: top;
  }

  /* 响应式按钮组样式 */
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 16px;
  }

  /* 小屏幕按钮文本显示调整 */
  @media screen and (width <= 768px) {
    .file-main {
      padding: 15px;
    }
    .idb-button-group :deep(.arco-btn) span {
      max-width: 80px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  /* 超小屏幕隐藏按钮文本 */
  @media screen and (width <= 576px) {
    .file-main {
      padding: 10px;
    }
    .idb-button-group :deep(.arco-btn) span {
      display: none;
    }
    .idb-button-group :deep(.arco-btn) svg {
      margin: 0;
    }
  }

  /* 极小屏幕 */
  @media screen and (width <= 480px) {
    .file-main {
      padding: 8px;
    }
  }
</style>
