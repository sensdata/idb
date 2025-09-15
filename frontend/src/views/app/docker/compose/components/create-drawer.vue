<template>
  <a-drawer
    v-model:visible="visible"
    :width="720"
    :title="t('app.docker.compose.create.title')"
    unmountOnClose
    :ok-loading="loading"
    @cancel="hide"
  >
    <a-form ref="formRef" :model="formState" layout="vertical">
      <a-form-item
        field="name"
        :label="t('app.docker.compose.create.form.name')"
      >
        <a-input
          v-model="formState.name"
          :placeholder="t('app.docker.compose.create.form.name.placeholder')"
        />
      </a-form-item>
    </a-form>

    <a-radio-group v-model="formState.active" type="button">
      <a-radio value="compose_content">
        {{ t('app.docker.compose.edit.compose_content') }}
      </a-radio>
      <a-radio value="env_content">
        {{ t('app.docker.compose.edit.env_content') }}
      </a-radio>
    </a-radio-group>

    <code-editor
      v-if="formState.active === 'compose_content'"
      v-model="formState.compose_content"
      :tab-size="4"
      :extensions="yamlExtensions"
      :autofocus="true"
      :indent-with-tab="true"
      :file="composeFile"
    />
    <code-editor
      v-else
      v-model="formState.env_content"
      :tab-size="4"
      :extensions="propExtension"
      :autofocus="true"
      :indent-with-tab="true"
      :file="envFile"
    />

    <template #footer>
      <div class="flex justify-end">
        <a-space>
          <a-button @click="hide">{{ t('common.cancel') }}</a-button>
          <a-button type="secondary" :loading="testing" @click="handleTest">
            {{ t('app.docker.compose.create.test') }}
          </a-button>
          <a-button type="primary" :loading="loading" @click="handleCreate">
            {{ t('app.docker.compose.create.submit') }}
          </a-button>
        </a-space>
      </div>
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, reactive, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue';
  import { Message } from '@arco-design/web-vue';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import { properties } from '@codemirror/legacy-modes/mode/properties';
  import CodeEditor from '@/components/code-editor/index.vue';
  import { testComposeApi, createComposeApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();

  const composeFile = computed(() => ({
    name: 'docker-compose.yml',
    path: '/tmp/docker-compose.yml',
  }));
  const envFile = computed(() => ({
    name: '.env',
    path: '/tmp/.env',
  }));

  const yamlExtensions = [StreamLanguage.define(yaml)];
  const propExtension = [StreamLanguage.define(properties)];

  const visible = ref(false);
  const loading = ref(false);
  const testing = ref(false);
  const formRef = ref<FormInstance>();

  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  const formState = reactive({
    name: '',
    active: 'compose_content',
    compose_content: '',
    env_content: '',
  });

  const validate = () => {
    if (!formState.name?.trim()) {
      Message.error(t('app.docker.compose.create.form.name.required'));
      return false;
    }
    if (!formState.compose_content?.trim()) {
      Message.error(t('app.docker.compose.edit.compose_content_required'));
      return false;
    }
    return true;
  };

  const handleTest = async () => {
    if (!validate()) return;
    try {
      testing.value = true;
      const result = await testComposeApi({
        name: formState.name,
        compose_content: formState.compose_content,
        env_content: formState.env_content,
      });
      if (result.success) {
        Message.success(t('common.message.operationSuccess'));
      } else {
        Message.error(result.error || t('common.message.operationError'));
      }
    } catch (e: any) {
      Message.error(e?.message || t('common.message.operationError'));
    } finally {
      testing.value = false;
    }
  };

  const handleCreate = async () => {
    if (!validate()) return;
    try {
      loading.value = true;
      await createComposeApi({
        name: formState.name,
        compose_content: formState.compose_content,
        env_content: formState.env_content,
      });
      Message.success(t('common.message.operationSuccess'));
      emit('success');
      // 不自动关闭，遵循“抽屉只在用户主动关闭时才关闭”的偏好
    } catch (e: any) {
      Message.error(e?.message || t('common.message.operationError'));
    } finally {
      loading.value = false;
    }
  };

  defineExpose({ show, hide });
</script>

<style scoped></style>
