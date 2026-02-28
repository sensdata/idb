<template>
  <div class="file-browser">
    <div class="header flex items-center justify-between">
      <span class="text-lg font-medium">{{
        t('components.file.fileSelector.title')
      }}</span>
      <icon-close
        class="text-gray-400 cursor-pointer hover:text-gray-600"
        @click="handleClose"
      />
    </div>
    <div class="current-path mb-4 px-4 py-2 bg-gray-50 rounded">
      <div class="flex items-center gap-1 overflow-hidden">
        <icon-left
          :class="['mr-2 cursor-pointer', canGoBack ? 'link' : 'text-gray-300']"
          @click="handleGoBack"
        />
        <div class="breadcrumb-wrapper overflow-hidden">
          <a-breadcrumb :max-count="4">
            <a-breadcrumb-item class="link" @click="handlePathClick('/')">
              <icon-home />
            </a-breadcrumb-item>
            <a-breadcrumb-item
              v-for="(part, index) in pathParts"
              :key="part.path"
              :class="[index === pathParts.length - 1 ? '' : 'link']"
              @click="handlePathClick(part.path)"
            >
              {{ part.name }}
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>
      </div>
    </div>

    <div class="operations mb-4 flex items-center gap-2">
      <a-button type="primary" size="small" @click="handleCreateFolder">
        <template #icon>
          <icon-folder-add />
        </template>
        {{ t('components.file.fileSelector.createFolder') }}
      </a-button>
      <a-button type="primary" size="small" @click="handleUploadFiles">
        <template #icon>
          <icon-upload />
        </template>
        {{ t('components.file.fileSelector.upload') }}
      </a-button>
    </div>

    <div class="search-box mb-4">
      <a-input-search
        v-model="searchQuery"
        :placeholder="t('components.file.fileSelector.searchPlaceholder')"
        :allow-clear="true"
      />
    </div>

    <div class="file-list h-[400px] overflow-y-auto">
      <a-spin
        :loading="isLoading"
        :tip="t('components.file.fileSelector.loading')"
        class="w-full"
      >
        <template v-if="!hasError">
          <a-empty
            v-if="filteredFileList.length === 0"
            :description="t('components.file.fileSelector.noData')"
          />
          <a-list v-else :bordered="false" size="small">
            <a-list-item
              v-for="file in filteredFileList"
              :key="file.path"
              class="cursor-pointer hover:bg-gray-50 rounded"
            >
              <div
                class="flex items-center w-full"
                @click="handleItemClick(file)"
              >
                <a-radio
                  :disabled="!isFileSelectable(file)"
                  :model-value="false"
                  class="mr-3"
                  @click.stop="handleFileSelect(file)"
                />
                <template v-if="file.is_dir">
                  <icon-folder class="mr-2 text-blue-500" />
                </template>
                <template v-else>
                  <icon-file class="mr-2 text-gray-500" />
                </template>
                <span class="flex-1 truncate">{{ file.name }}</span>
              </div>
            </a-list-item>
          </a-list>
        </template>
        <a-result
          v-else
          status="error"
          :title="t('components.file.fileSelector.error')"
        />
      </a-spin>
    </div>
  </div>
  <create-folder-drawer
    ref="createFolderDrawerRef"
    :host="hostId"
    @ok="handleOperationSuccess"
  />
  <upload-files-drawer
    ref="uploadFilesDrawerRef"
    :host="hostId"
    @ok="handleOperationSuccess"
  />
</template>

<script setup lang="ts">
  import { ref, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { getFileListApi } from '@/api/file';
  import { useHostStore } from '@/store';
  import CreateFolderDrawer from '@/components/file/create-folder-drawer/index.vue';
  import UploadFilesDrawer from '@/components/file/upload-files-drawer/index.vue';
  import { FileSelectorItem as FileItem, FileSelectType } from './types';

  const hostStore = useHostStore();

  interface Props {
    initialPath?: string;
    type?: FileSelectType | string;
    host?: number;
  }

  interface Emits {
    (e: 'select', file: FileItem): void;
    (e: 'cancel'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    initialPath: '/',
    type: FileSelectType.ALL,
  });

  const hostId = computed(
    () => props.host ?? hostStore.currentId ?? hostStore.defaultId
  );

  const emit = defineEmits<Emits>();
  const { t } = useI18n();

  const currentPath = ref<string>(props.initialPath);
  const searchQuery = ref<string>('');
  const fileList = ref<FileItem[]>([]);
  const isLoading = ref<boolean>(false);
  const hasError = ref<boolean>(false);

  const filteredFileList = computed(() => {
    return fileList.value.filter((file) =>
      file.name.toLowerCase().includes(searchQuery.value.toLowerCase())
    );
  });
  const canGoBack = computed(() => currentPath.value !== '/');
  const normalizedSelectType = computed(() => {
    const rawType = String(props.type || '').toLowerCase();
    if (rawType === 'directory') return FileSelectType.DIR;
    if (rawType === FileSelectType.FILE) return FileSelectType.FILE;
    if (rawType === FileSelectType.DIR) return FileSelectType.DIR;
    return FileSelectType.ALL;
  });

  const pathParts = computed(() => {
    const parts = currentPath.value.split('/').filter(Boolean);
    return parts.map((part, index) => ({
      name: part,
      path: '/' + parts.slice(0, index + 1).join('/'),
    }));
  });

  const isFileSelectable = (file: FileItem): boolean => {
    switch (normalizedSelectType.value) {
      case FileSelectType.FILE:
        return !file.is_dir;
      case FileSelectType.DIR:
        return file.is_dir;
      case FileSelectType.ALL:
      default:
        return true;
    }
  };

  const getParentPath = (path: string): string => {
    if (path === '/') return '/';
    const parts = path.split('/').filter(Boolean);
    parts.pop();
    return '/' + parts.join('/');
  };

  const loadFileList = async (path: string): Promise<void> => {
    isLoading.value = true;
    hasError.value = false;
    try {
      const data = await getFileListApi({
        host: hostId.value,
        page: 1,
        page_size: 200,
        show_hidden: true,
        path,
      });
      fileList.value = data.items;
    } catch (error) {
      hasError.value = true;
      console.error('Failed to load file list:', error);
    } finally {
      isLoading.value = false;
    }
  };

  const handleItemClick = async (file: FileItem): Promise<void> => {
    if (file.is_dir) {
      currentPath.value = file.path;
      await loadFileList(file.path);
    } else if (isFileSelectable(file)) {
      emit('select', file);
    }
  };

  const handleFileSelect = (file: FileItem): void => {
    if (isFileSelectable(file)) {
      emit('select', file);
    }
  };

  const handleClose = (): void => {
    emit('cancel');
  };

  const handleGoBack = async (): Promise<void> => {
    if (canGoBack.value) {
      const parentPath = getParentPath(currentPath.value);
      currentPath.value = parentPath;
      await loadFileList(parentPath);
    }
  };

  const handlePathClick = async (path: string): Promise<void> => {
    if (path === currentPath.value) {
      return;
    }
    currentPath.value = path;
    searchQuery.value = '';
    await loadFileList(path);
  };

  const createFolderDrawerRef = ref();
  const uploadFilesDrawerRef = ref();

  const handleCreateFolder = () => {
    createFolderDrawerRef.value?.show();
    createFolderDrawerRef.value?.setData({
      pwd: currentPath.value,
    });
  };

  const handleUploadFiles = () => {
    uploadFilesDrawerRef.value?.show();
    uploadFilesDrawerRef.value?.setData({
      directory: currentPath.value,
    });
  };

  const handleOperationSuccess = async () => {
    await loadFileList(currentPath.value);
  };

  onMounted(() => {
    loadFileList(currentPath.value);
  });
</script>

<style lang="less" scoped>
  .file-browser {
    width: 360px;
  }

  .file-list {
    border: 1px solid var(--color-neutral-3);
    border-radius: 4px;
    :deep(.arco-list-small .arco-list-item) {
      padding: 4px 12px;
    }
    :deep(.arco-list-split .arco-list-item:not(:last-child)) {
      border-bottom-style: dashed;
    }
  }

  .header {
    padding: 8px 16px;
    margin: -16px -16px 8px -16px;
    border-bottom: 1px solid var(--color-neutral-3);
  }

  .current-path {
    background-color: var(--color-fill-2);
  }

  .breadcrumb-wrapper {
    :deep(.arco-breadcrumb) {
      white-space: nowrap;
    }
    :deep(.arco-breadcrumb-item) {
      &.link {
        color: rgb(var(--link-6));
        cursor: pointer;
      }
    }
    :deep(.arco-breadcrumb-item-separator) {
      margin: 0;
    }
  }

  .operations {
    :deep(.arco-btn) {
      .arco-icon {
        margin-right: 4px;
      }
    }
  }
</style>
