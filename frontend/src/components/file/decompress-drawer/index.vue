<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.decompressDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @before-ok="handleBeforeOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules">
      <a-form-item
        field="file"
        :label="$t('components.file.decompressDrawer.file')"
      >
        <span>{{ fileName }}</span>
      </a-form-item>
      <a-form-item
        field="dst"
        :label="$t('components.file.decompressDrawer.dstType')"
      >
        <div>
          <a-radio-group v-model="formState.dst_type" type="button">
            <a-radio value="current">{{
              $t('components.file.decompressDrawer.currentDir')
            }}</a-radio>
            <a-radio value="custom">{{
              $t('components.file.decompressDrawer.customDir')
            }}</a-radio>
          </a-radio-group>
          <file-selector
            v-if="formState.dst_type === 'custom'"
            ref="fileSelectorRef"
            v-model="formState.dst"
            :initial-path="formState.dst"
            class="mt-2"
            type="directory"
            :placeholder="$t('components.file.decompressDrawer.dstPlaceholder')"
          />
        </div>
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { decompressFilesApi } from '@/api/file';
  import FileSelector from '@/components/file/file-selector/index.vue';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const fileSelectorRef = ref<InstanceType<typeof FileSelector>>();
  const files = ref<FileInfoEntity[]>([]);
  const formState = reactive({
    dst_type: 'current',
    dst: '',
  });

  const fileName = computed(() => {
    return files.value[0]?.name || '';
  });

  const rules = computed(() => ({
    dst: [
      {
        required: true,
        validator: (value: string, callback: any) => {
          if (formState.dst_type === 'custom') {
            if (!formState.dst) {
              callback(t('components.file.decompressDrawer.dstRequired'));
            } else {
              callback();
            }
          } else {
            callback();
          }
        },
        trigger: 'blur',
      },
    ],
  }));

  // 切换目录类型时清除验证信息
  watch(
    () => formState.dst_type,
    () => {
      formRef.value?.clearValidate();
    }
  );

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleBeforeOk = async () => {
    if (await validate()) {
      try {
        showLoading();
        await decompressFilesApi({
          path: files.value[0].path,
          dst:
            formState.dst_type === 'current'
              ? files.value[0].path.split('/').slice(0, -1).join('/')
              : formState.dst,
        });
        Message.success(t('components.file.decompressDrawer.success'));
        emit('ok');
        return true;
      } catch (err: any) {
        Message.error(err);
      } finally {
        hideLoading();
      }
    }
    return false;
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const setFiles = (selectedFiles: FileInfoEntity[]) => {
    files.value = selectedFiles;
    if (selectedFiles.length > 0) {
      formState.dst = selectedFiles[0].path.split('/').slice(0, -1).join('/');
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
    max-height: 400px;
    padding: 8px;
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
    padding: 5px 8px;
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
