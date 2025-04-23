<template>
  <div class="file-main">
    <div class="action-wrap mb-4">
      <a-button-group class="idb-button-group">
        <a-button @click="$emit('back')">
          <icon-left />
          <span class="ml-2">{{ $t('app.file.list.action.back') }}</span>
        </a-button>
        <template v-if="!selected.length">
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
        </template>
        <template v-else>
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
        </template>
        <a-button @click="$emit('terminal')">
          <icon-code-square />
          <span class="ml-2">{{ $t('app.file.list.action.terminal') }}</span>
        </a-button>
      </a-button-group>
      <a-button-group v-if="pasteVisible" class="idb-button-group ml-4">
        <a-button @click="$emit('paste')">
          <icon-paste />
          <span class="ml-2">
            {{ $t('app.file.list.action.paste', { count: selected.length }) }}
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
      has-batch
      row-key="path"
      @selected-change="$emit('selectedChange', $event)"
    >
      <template #leftActions>
        <a-checkbox
          :modelValue="showHidden"
          @update:model-value="$emit('update:showHidden', $event)"
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
          <folder-icon v-if="record.is_dir" />
          <file-icon v-else />
          <span class="color-primary cursor-pointer min-w-0 flex-1 truncate">{{
            record.name
          }}</span>
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
  import { ref, defineProps, defineEmits, GlobalComponents } from 'vue';
  import { debounce } from 'lodash';
  import { getFileListApi } from '@/api/file';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import CompressionIcon from '@/assets/icons/compression.svg';
  import DecompressionIcon from '@/assets/icons/decompression.svg';
  import { FileItem } from '../types/file-item';

  interface TableColumn {
    dataIndex: string;
    title: string;
    width?: number;
    ellipsis?: boolean;
    slotName?: string;
    render?: (data: any) => any;
    align?: 'left' | 'center' | 'right';
  }

  defineProps({
    params: {
      type: Object,
      required: true,
    },
    columns: {
      type: Array as () => TableColumn[],
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    showHidden: {
      type: Boolean,
      default: false,
    },
    selected: {
      type: Array as () => FileItem[],
      default: () => [],
    },
    pasteVisible: {
      type: Boolean,
      default: false,
    },
    decompressVisible: {
      type: Boolean,
      default: false,
    },
  });

  const emit = defineEmits([
    'update:showHidden',
    'copy',
    'cut',
    'paste',
    'create',
    'upload',
    'compress',
    'decompress',
    'delete',
    'terminal',
    'clearSelected',
    'selectedChange',
    'itemSelect',
    'itemDoubleClick',
    'modifyMode',
    'modifyOwner',
    'operation',
    'back',
    'reload',
  ]);

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  // Handle item selection with debounce to avoid conflicts with double-click
  const handleItemSelect = debounce((record: FileItem) => {
    emit('itemSelect', record);
  }, 250);

  // Handle double click on a file/folder
  const handleItemDoubleClick = (record: FileItem) => {
    // Cancel the pending single-click event
    handleItemSelect.cancel();
    emit('itemDoubleClick', record);
  };

  // Handle dropdown menu operations
  const handleOperation = (key: string, record: FileItem) => {
    emit('operation', key, record);
  };

  // Handle create folder/file dropdown selection
  const handleCreate = (key: string) => {
    emit('create', key);
  };

  // Public methods exposed for parent component
  const load = (params: any) => {
    gridRef.value?.load(params);
  };

  const reload = () => {
    gridRef.value?.reload();
  };

  const clearSelected = () => {
    gridRef.value?.clearSelected();
  };

  const getData = () => {
    return gridRef.value?.getData() || [];
  };

  // Expose methods to parent component
  defineExpose({
    load,
    reload,
    clearSelected,
    getData,
  });
</script>

<style scoped>
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
