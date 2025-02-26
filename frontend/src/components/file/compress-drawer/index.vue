<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.compressDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item
        field="type"
        :label="$t('components.file.compressDrawer.type')"
      >
        <a-select v-model="formState.type" class="w-32">
          <a-option value="tar.gz">tar.gz</a-option>
        </a-select>
      </a-form-item>
      <a-form-item
        field="name"
        :label="$t('components.file.compressDrawer.name')"
      >
        <a-input
          v-model="formState.name"
          :placeholder="$t('components.file.compressDrawer.namePlaceholder')"
        >
          <template #append> .{{ formState.type }} </template>
        </a-input>
      </a-form-item>
      <a-form-item
        field="replace"
        :label="$t('components.file.compressDrawer.replace')"
      >
        <a-switch v-model="formState.replace" />
      </a-form-item>
      <a-form-item
        field="files"
        :label="$t('components.file.compressDrawer.files')"
      >
        <div class="file-list">
          <div v-for="file in files" :key="file.path" class="file-item">
            <div class="file-item-icon">
              <FolderIcon v-if="file.is_dir" />
              <FileIcon v-else />
            </div>
            <div class="file-item-name">
              {{ file.name }}
            </div>
          </div>
        </div>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { compressFilesApi } from '@/api/file';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const files = ref<FileInfoEntity[]>([]);
  const formState = reactive({
    name: '',
    type: 'tar.gz',
    replace: false,
  });

  const rules = {
    name: [
      {
        required: true,
        message: t('components.file.compressDrawer.nameRequired'),
      },
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (value && !/^[a-zA-Z0-9-_.]+$/.test(value)) {
            callback(t('components.file.compressDrawer.nameInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    type: [
      {
        required: true,
        message: t('components.file.compressDrawer.typeRequired'),
      },
    ],
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      await compressFilesApi({
        name: formState.name + '.' + formState.type,
        files: files.value.map((f) => f.path),
        dst: files.value[0].path.split('/').slice(0, -1).join('/'),
        type: formState.type,
        replace: formState.replace,
      });
      Message.success(t('components.file.compressDrawer.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const setFiles = (selectedFiles: FileInfoEntity[]) => {
    files.value = selectedFiles;
    if (selectedFiles.length === 1) {
      formState.name = selectedFiles[0].name.split('.')[0];
    } else if (selectedFiles.length > 1) {
      const paths = selectedFiles[0].path.split('/');
      paths.pop();
      formState.name = paths[paths.length - 1] || 'compressed-' + Date.now();
    }
  };

  defineExpose({
    show,
    hide,
    setFiles,
  });
</script>

<style scoped>
  .file-list {
    width: 100%;
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid var(--color-border);
    border-radius: 4px;
  }

  .file-item {
    display: flex;
    flex: 1;
    align-items: center;
    width: 100%;
    min-width: 0;
    padding: 5px 12px;
    border-bottom: 1px solid var(--color-border);
  }

  .file-item:last-child {
    border-bottom: none;
  }

  .file-item-icon {
    display: flex;
    align-items: center;
    height: 100%;
    vertical-align: top;
  }

  .file-item-icon svg {
    width: 14px;
    height: 14px;
  }

  .file-item-name {
    flex: 1;
    min-width: 0;
    margin-left: 8px;
    overflow: hidden;
    font-size: 14px;
    line-height: 22px;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
</style>
