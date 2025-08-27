<template>
  <a-drawer
    v-model:visible="visible"
    :width="720"
    :title="t('app.docker.image.list.action.build')"
    :ok-loading="loading"
    unmount-on-close
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="t('app.docker.image.form.name')"
        :rules="[
          {
            required: true,
            message: t('app.docker.image.form.name.required'),
          },
        ]"
      >
        <a-input
          v-model="formState.name"
          :placeholder="t('app.docker.image.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item field="mode">
        <a-radio-group v-model="formState.mode" type="button">
          <a-radio value="edit">{{
            t('app.docker.image.form.mode.edit')
          }}</a-radio>
          <a-radio value="file">{{
            t('app.docker.image.form.mode.file')
          }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <a-form-item
        v-if="formState.mode === 'edit'"
        field="docker_file_content"
        :label="t('app.docker.image.form.docker_file_content')"
        :rules="[
          {
            required: true,
            message: t('app.docker.image.form.docker_file_content.required'),
          },
        ]"
      >
        <div style="width: 100%; height: 400px">
          <code-editor
            v-model="formState.docker_file_content"
            :tab-size="4"
            :extensions="extensions"
            :autofocus="true"
            :indent-with-tab="true"
            :file="dockerFileObj"
          />
        </div>
      </a-form-item>
      <a-form-item
        v-if="formState.mode === 'file'"
        field="docker_file"
        :label="t('app.docker.image.form.docker_file')"
        :rules="[
          {
            required: true,
            message: t('app.docker.image.form.docker_file.required'),
          },
        ]"
      >
        <file-selector
          v-model="formState.docker_file"
          :initial-path="formState.docker_file"
          type="file"
          :placeholder="t('app.docker.image.form.docker_file.placeholder')"
        />
      </a-form-item>

      <a-form-item field="tags" :label="t('app.docker.image.form.tags')">
        <a-input-tag
          v-model="formState.tags"
          :placeholder="t('app.docker.image.form.tags.placeholder')"
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, reactive, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import { buildImageApi } from '@/api/docker';
  import { StreamLanguage } from '@codemirror/language';
  import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
  import CodeEditor from '@/components/code-editor/index.vue';
  import FileSelector from '@/components/file/file-selector/index.vue';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();

  // 创建 Dockerfile 文件对象
  const dockerFileObj = computed(() => ({
    name: 'Dockerfile',
    path: '/tmp/Dockerfile',
  }));

  const extensions = [StreamLanguage.define(dockerFile)];
  const formState = reactive({
    name: '',
    mode: 'edit',
    docker_file: '',
    docker_file_content: '',
    tags: [] as string[],
  });
  const rules = {
    name: [
      {
        required: true,
        message: t('app.docker.image.form.name.required'),
      },
    ],
    docker_file: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (formState.mode === 'file' && !value) {
            callback(t('app.docker.image.form.docker_file.required'));
          } else {
            callback();
          }
        },
      },
    ],
    docker_file_content: [
      {
        validator: (value: string, callback: (error?: string) => void) => {
          if (formState.mode === 'edit' && !value) {
            callback(t('app.docker.image.form.docker_file_content.required'));
          } else {
            callback();
          }
        },
      },
    ],
  };

  const show = () => {
    visible.value = true;
    formState.mode = 'edit';
    formState.docker_file = '';
    formState.docker_file_content = '';
    formState.name = '';
    formState.tags = [];
    formRef.value?.resetFields();
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };
  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };
  const onBeforeOk = async () => {
    if (!(await validate())) {
      return false;
    }

    try {
      loading.value = true;
      await buildImageApi({
        docker_file:
          formState.mode === 'file'
            ? formState.docker_file
            : formState.docker_file_content,
        from: formState.mode,
        name: formState.name,
        tags: formState.tags,
      });
      Message.success(t('app.docker.image.build.success'));
      emit('success');
      hide();
    } catch (e: any) {
      if (e?.message) {
        Message.error(e.message);
      }
    } finally {
      loading.value = false;
    }

    return true;
  };
  const onCancel = hide;

  defineExpose({ show });
</script>

<style scoped></style>
