<template>
  <div class="file-main">
    <div class="toolbar">
      <div class="toolbar-left">
        <a-button-group class="idb-button-group">
          <a-button @click="$emit('back')">
            <icon-left />
          </a-button>
          <a-dropdown
            position="bl"
            @select="(key: string) => handleCreate(key)"
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
            <span class="ml-2">{{
              $t('app.file.list.action.decompress')
            }}</span>
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
                $t('app.file.list.action.paste', {
                  count: props.clipboardCount || 0,
                })
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

      <!-- Pin current directory button on the right -->
      <div class="toolbar-right">
        <a-button
          :type="isCurrentDirectoryPinned ? 'primary' : 'outline'"
          size="small"
          :loading="loadingFavorites"
          :class="{
            'favorite-button': true,
            'favorite-button-active': isCurrentDirectoryPinned,
          }"
          @click="toggleCurrentDirectoryFavorite"
        >
          <icon-pushpin />
          <span class="ml-1 favorite-button-text">
            {{
              $t(
                isCurrentDirectoryPinned
                  ? 'app.file.list.action.unfavoriteCurrent'
                  : 'app.file.list.action.favoriteCurrent'
              )
            }}
          </span>
        </a-button>
      </div>
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
          <icon-link v-if="record.is_symlink" style="color: var(--idblue-6)" />
          <span
            class="cursor-pointer min-w-0 flex-1 truncate"
            style="color: var(--color-text-1)"
          >
            {{ record.name }}
            <template v-if="record.is_symlink">
              <span class="ml-1" style="color: var(--color-text-3)">→</span>
              <span class="ml-1 italic" style="color: var(--color-text-3)">{{
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
          @select="(key: string) => handleOperation(key, record)"
        >
          <a-button type="text" class="operation-button">
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
  import { ref, GlobalComponents, watch, computed, onMounted } from 'vue';
  import { debounce } from 'lodash';
  import { Message } from '@arco-design/web-vue';
  import { IconPushpin } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import { getFileListApi } from '@/api/file';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import CompressionIcon from '@/assets/icons/compression.svg';
  import DecompressionIcon from '@/assets/icons/decompression.svg';
  import { FileItem } from '@/components/file/file-editor-drawer/types';
  import { useLogger } from '@/composables/use-logger';
  import { usePinnedDirectories } from '../composables/use-pinned-directories';

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
    clipboardCount?: number;
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
  const { t } = useI18n();

  // 初始化日志
  const { logInfo, logWarn } = useLogger('FileMainView');

  // 收藏目录状态
  const {
    isPinned,
    pinDirectory,
    unpinDirectory,
    loadFavorites,
    loadingFavorites,
  } = usePinnedDirectories();

  // 直接使用props中的状态
  const selectedItems = computed(() => props.selected || []);

  // Get current directory path from params
  const currentDirectoryPath = computed(() => {
    return props.params?.path || '/';
  });

  // Check if current directory is pinned
  const isCurrentDirectoryPinned = computed(() => {
    return isPinned(currentDirectoryPath.value);
  });

  // Toggle favorite status for current directory
  const toggleCurrentDirectoryFavorite = async () => {
    const path = currentDirectoryPath.value;
    try {
      if (isCurrentDirectoryPinned.value) {
        await unpinDirectory(path);
      } else {
        // Extract directory name from path for display
        const name = path.split('/').filter(Boolean).pop() || 'root';
        await pinDirectory(path, name);
      }
    } catch (error: any) {
      Message.error(error?.message || t('common.request.failed'));
    }
  };

  onMounted(() => {
    loadFavorites().catch((error) => {
      logWarn('Failed to load favorite directories in main view:', error);
    });
  });

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

<style scoped lang="less">
  @import url('@/assets/style/mixin.less');

  .file-main {
    display: flex;
    flex: 1;
    flex-direction: column;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 20px;
    overflow: hidden;
    background-color: var(--color-bg-1);
  }

  :deep(.idb-table) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  :deep(.arco-table) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  /* Keep table body inside available panel height instead of viewport math */
  :deep(.arco-table-container) {
    flex: 1;
    min-height: 0;
    overflow: hidden;

    /* Apply custom scrollbar styling to table container */
    .custom-scrollbar();
  }

  /* Table body scroll area */
  :deep(.arco-scrollbar-container.arco-table-content) {
    max-height: none;
    overflow-y: auto;

    /* Apply custom scrollbar styling */
    .custom-scrollbar();
  }

  :deep(.arco-table-tr-empty .arco-table-td) {
    height: clamp(220px, 38vh, 360px);
    padding-top: 0;
    padding-bottom: 0;
  }

  :deep(.arco-empty) {
    color: var(--color-text-3);
  }

  .name-cell svg {
    width: 14px;
    height: 14px;
    margin-right: 8px;
    vertical-align: top;
  }

  /* 操作列按钮样式 */
  .operation-button {
    padding-left: 0 !important;
  }

  /* 响应式按钮组样式 */
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  .toolbar-left {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 0.625rem;
  }

  .toolbar-right {
    display: flex;
    flex-shrink: 0;
  }

  .favorite-button {
    height: 32px;
    padding: 0 0.625rem !important;
    font-size: 0.8125rem;
    transition: all 0.2s ease;
  }

  .favorite-button :deep(svg) {
    font-size: 0.875rem;
  }

  .favorite-button-active {
    box-shadow: 0 0 0 1px rgb(var(--primary-3)) inset;
  }

  .favorite-button-text {
    margin-left: 0.25rem;
    font-size: 0.8125rem;
    font-weight: 500;
  }

  /* 小屏幕按钮文本显示调整 */
  @media screen and (width <= 768px) {
    .file-main {
      padding: 0.9375rem;
    }
    .toolbar {
      gap: 0.5rem;
    }
    .toolbar-left {
      gap: 0.5rem;
    }
    .idb-button-group :deep(.arco-btn) span {
      max-width: 5rem;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .favorite-button-text {
      max-width: 3rem;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  /* 超小屏幕隐藏按钮文本 */
  @media screen and (width <= 576px) {
    .file-main {
      padding: 0.625rem;
    }
    .toolbar {
      flex-direction: column;
      gap: 0.5rem;
      align-items: stretch;
    }
    .toolbar-right {
      justify-content: flex-end;
      width: 100%;
    }
    .idb-button-group :deep(.arco-btn) span {
      display: none;
    }
    .idb-button-group :deep(.arco-btn) svg {
      margin: 0;
    }
    .favorite-button-text {
      display: none;
    }
    .favorite-button :deep(svg) {
      margin: 0;
    }
  }

  /* 极小屏幕 */
  @media screen and (width <= 480px) {
    .file-main {
      padding: 0.5rem;
    }
  }
</style>
